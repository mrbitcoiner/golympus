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

type appHandler = func(w http.ResponseWriter, r *http.Request) error

func main() {
	fmt.Fprintln(os.Stderr, "golympus", version, "by theBitcoinheiro")
	pf := NewPriceFetcher()
	ff := NewFeerateFetcher(cfg.BtcUrl, cfg.BtcUser, cfg.BtcPassword)
	lr := NewLnRouter(cfg.LnNetwork, cfg.LnAddress)
	srv := newServer(pf, ff, lr)

	http.HandleFunc("POST /rates/get", httpErrMdw(srv.ratesHandler))
	http.HandleFunc("POST /router/routesplus", httpErrMdw(srv.routesplusHandler))
	//http.HandleFunc("POST /router/routesplus", srv.hardcodedRoutesPlus)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s request on %s\n", r.Method, r.URL.Path)
	})

	fmt.Fprintln(os.Stderr, "to listen on", cfg.ListenAddr)
	http.ListenAndServe(cfg.ListenAddr, nil)
}

func httpErrMdw(fn appHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err != nil {
			log.Println("error middleware:\n", err)
			w.WriteHeader(500)
		}
	}
}
