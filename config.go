package main

import (
	"master.private/bstd.git/util"
	"github.com/joho/godotenv"
)

var cfg config

type config struct {
	ListenAddr string
	BtcUrl string
	BtcUser string
	BtcPassword string
}

func init() {
	godotenv.Load()
	cfg = config{
		ListenAddr: util.EnvOrDefault("LISTEN_ADDR", "0.0.0.0:8088"),
		BtcUrl: util.MustEnv("BTC_URL"),
		BtcUser: util.MustEnv("BTC_USER"),
		BtcPassword: util.MustEnv("BTC_PASSWORD"),
	}
}
