package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"master.private/bstd.git/stackerr"
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
	srv := newServer(pf)
	listenAddr := "0.0.0.0:8085"
	http.HandleFunc("POST /rates/get", srv.ratesHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("request on /")
	})
	fmt.Fprintln(os.Stderr, "golympus to listen on", listenAddr)
	http.ListenAndServe(listenAddr, nil)
}

type PriceFetcher interface {
	Fetch(symbols ...Symbol) (map[Symbol]float64, error)
}

type server struct {
	pf PriceFetcher
}

func newServer(pf PriceFetcher) *server {
	return &server {
		pf: pf,
	}
}
 
func (s *server) ratesHandlerOld(w http.ResponseWriter, _ *http.Request) {
	log.Println("request on POST /rates/get")
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	fmt.Fprintf(
		w, `["ok",[{"12":0.0,"8":0.0,"4":0.0,"11":0.0,"9":0.0,"5":0.0,"10":0.0,`+
		`"6":0.0,"2":0.0,"7":0.0,"3":0.0},`+
		`{"usd":105000.05000,"eur":101000.10000,"jpy":14436600.000,`+
		`"cny":50000.10000,"brl":580000.20000}]]`,
	) 
}

func (s *server) ratesHandler(w http.ResponseWriter, _ *http.Request) {
	log.Println("request on POST /rates/get")
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	prices, err := s.pf.Fetch(USD, EUR, JPY, CNY, BRL)
	if err != nil {
		log.Println(stackerr.Wrap(err))
		w.WriteHeader(500)
		return
	}
	res := []interface{}{
		"ok",
		[]map[string]float64{
			{
				"3": 0.0,
				"6": 0.0,
			},
			prices,
		},
	}
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Println(stackerr.Wrap(err))
	}
}
