package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var version string

func must(err error) {
	if err == nil {
		return
	}
	panic(err)
}

func main() {
	fmt.Fprintln(os.Stderr, "golympus", version, "by theBitcoinheiro")
	pf := NewPriceFetcher()
	ff := NewFeerateFetcher(cfg.BtcUrl, cfg.BtcUser, cfg.BtcPassword)
	lr := NewLnRouter(cfg.LnNetwork, cfg.LnAddress)
	srv := newServer(pf, ff, lr)

	http.HandleFunc("POST /rates/get", srv.ratesHandler)
	http.HandleFunc("POST /router/routesplus", srv.routesplusHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s request on %s\n", r.Method, r.URL.Path)
	})

	fmt.Fprintln(os.Stderr, "to listen on", cfg.ListenAddr)
	http.ListenAndServe(cfg.ListenAddr, nil)
}
