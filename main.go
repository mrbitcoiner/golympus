package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"master.private/bstd.git/stackerr"
	"master.private/bstd.git/util"
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
	ff := NewFeerateFetcher(
		util.MustEnv("BTC_URL"),
		util.MustEnv("BTC_USER"),
		util.MustEnv("BTC_PASSWORD"),
	)
	srv := newServer(pf, ff)
	listenAddr := util.EnvOrDefault("LISTEN_ADDR", "0.0.0.0:8088")
	http.HandleFunc("POST /rates/get", srv.ratesHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("request on /")
	})
	fmt.Fprintln(os.Stderr, "golympus to listen on", listenAddr)
	http.ListenAndServe(listenAddr, nil)
}

type PriceFetcher interface {
	FetchPrice(symbols ...Symbol) (map[Symbol]float64, error)
}

type FeerateFetcher interface {
	FetchFeerate(nBlockTarget ...int32) (map[int32]float64, error)
}

type server struct {
	pf PriceFetcher
	ff FeerateFetcher
}

func newServer(pf PriceFetcher, ff FeerateFetcher) *server {
	return &server {
		pf: pf,
		ff: ff,
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

	feerates, err := s.ff.FetchFeerate(1,2,3,4,5,6,7,8,9,10,11,12)
	if err != nil {
		log.Println(stackerr.Wrap(err))
		w.WriteHeader(500)
		return
	}
	feeratesResult := map[string]float64{}
	for k, v := range feerates {
		asStr := strconv.FormatInt(int64(k), 10)
		feeratesResult[asStr] = v
	}

	prices, err := s.pf.FetchPrice(USD, EUR, JPY, CNY, BRL)
	if err != nil {
		log.Println(stackerr.Wrap(err))
		w.WriteHeader(500)
		return
	}

	res := []interface{}{
		"ok",
		[]map[string]float64{
			feeratesResult,
			prices,
		},
	}
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Println(stackerr.Wrap(err))
	}
}
