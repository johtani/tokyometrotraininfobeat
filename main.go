package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/johtani/tokyometrotraininfobeat/beater"
)

func main() {
	err := beat.Run("tokyometrotraininfobeat", "", beater.New())
	if err != nil {
		os.Exit(1)
	}
}
