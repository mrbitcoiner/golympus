package main

import (
	"github.com/joho/godotenv"
	"master.private/bstd.git/util"
)

var cfg config

type config struct {
	ListenAddr  string
	BtcUrl      string
	BtcUser     string
	BtcPassword string
	LnNetwork   string
	LnAddress   string
}

func init() {
	godotenv.Load()
	cfg = config{
		ListenAddr:  util.EnvOrDefault("LISTEN_ADDR", "0.0.0.0:8088"),
		BtcUrl:      util.MustEnv("BTC_URL"),
		BtcUser:     util.MustEnv("BTC_USER"),
		BtcPassword: util.MustEnv("BTC_PASSWORD"),
		LnNetwork:   util.MustEnv("LN_NETWORK"),
		LnAddress:   util.MustEnv("LN_ADDRESS"),
	}
}
