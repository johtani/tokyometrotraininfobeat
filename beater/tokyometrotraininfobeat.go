package beater

import (
	"fmt"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/johtani/tokyometrotraininfobeat/config"
	"net/http"
	"io/ioutil"
	"github.com/bitly/go-simplejson"
)

type Tokyometrotraininfobeat struct {
	beatConfig *config.Config
	done       chan struct{}
	period     time.Duration
	events     publisher.Client
	uri        string
}

// Creates beater
func New() *Tokyometrotraininfobeat {
	return &Tokyometrotraininfobeat{
		done: make(chan struct{}),
	}
}

/// *** Beater interface methods ***///

func (bt *Tokyometrotraininfobeat) Config(b *beat.Beat) error {

	// Load beater beatConfig
	err := b.RawConfig.Unpack(&bt.beatConfig)
	if err != nil {
		return fmt.Errorf("Error reading config file: %v", err)
	}

	if bt.beatConfig.Tokyometrotraininfobeat.Token == "" {
		return fmt.Errorf("Must set 'token'")
	}

	return nil
}

func (bt *Tokyometrotraininfobeat) Setup(b *beat.Beat) error {

	// Setting default period if not set
	if bt.beatConfig.Tokyometrotraininfobeat.Period == "" {
		bt.beatConfig.Tokyometrotraininfobeat.Period = "1s"
	}

	bt.events = b.Publisher.Connect()

	var err error
	bt.period, err = time.ParseDuration(bt.beatConfig.Tokyometrotraininfobeat.Period)
	if err != nil {
		return err
	}

	bt.uri = bt.beatConfig.Tokyometrotraininfobeat.Uri + bt.beatConfig.Tokyometrotraininfobeat.Token

	return nil
}

func (bt *Tokyometrotraininfobeat) Run(b *beat.Beat) error {
	logp.Info("tokyometrotraininfobeat is running! Hit CTRL-C to stop it.")

	ticker := time.NewTicker(bt.period)
	logp.Info(bt.uri)
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}
		logp.Info("after select")
		resp, _ := http.Get(bt.uri)
		logp.Info("get response : %v",resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		rootjson, err3 := simplejson.NewJson(body)
		if err3 != nil { return err3}

		for i := 0; i < len(rootjson.MustArray()); i++ {
			event := common.MapStr{
				"@timestamp":  getParsedTime("dc:date", rootjson.GetIndex(i)),
				"type":        b.Name,
				"operator":    getOperator(rootjson.GetIndex(i)),
				"line":        getRailway(rootjson.GetIndex(i)),
				"status":      getStatus(rootjson.GetIndex(i)),
				"reason":      getReason(rootjson.GetIndex(i)),
				"origin_date": getParsedTime("odpt:timeOfOrigin", rootjson.GetIndex(i)),
			}
			bt.events.PublishEvent(event)
		}
		logp.Info("Event sent")
	}
}

func getParsedTime(attr string, entry *simplejson.Json) common.Time {
	timestring := entry.Get(attr).MustString()
	logp.Info("time:["+timestring+"]")
	secondtimestamp, err := time.Parse("2006-01-02T15:04:05-07:00", timestring)
	if err != nil {
		logp.Info("errororor")
	}
	return common.Time(secondtimestamp)
}

func getStatus(entry *simplejson.Json) string {
	return entry.Get("odpt:trainInformationStatus").MustString()
}

func getOperator(entry *simplejson.Json) string {
	return entry.Get("odpt:operator").MustString()
}

func getRailway(entry *simplejson.Json) string {
	return entry.Get("odpt:railway").MustString()
}

func getReason(entry *simplejson.Json) string {
	return entry.Get("odpt:trainInformationText").MustString()
}

func (bt *Tokyometrotraininfobeat) Cleanup(b *beat.Beat) error {
	return nil
}

func (bt *Tokyometrotraininfobeat) Stop() {
	close(bt.done)
}
