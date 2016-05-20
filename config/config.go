// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

type Config struct {
	Tokyometrotraininfobeat TokyometrotraininfobeatConfig
}

type TokyometrotraininfobeatConfig struct {
	Period string `config:"period"`
	Token string `config:"token"`
	Uri string `config:uri`
}
