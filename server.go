package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"master.private/bstd.git/stackerr"
)

type server struct {
	pf PriceFetcher
	ff FeerateFetcher
	rf RouteFinder
}

func newServer(pf PriceFetcher, ff FeerateFetcher, rf RouteFinder) *server {
	return &server{
		pf: pf,
		ff: ff,
		rf: rf,
	}
}

func (s *server) ratesHandler(w http.ResponseWriter, _ *http.Request) error {
	log.Println("request on POST /rates/get")
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")

	feerates, err := s.ff.FetchFeerate(1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12)
	if err != nil {
		return stackerr.Wrap(err)
	}
	feeratesResult := map[string]float64{}
	for k, v := range feerates {
		asStr := strconv.FormatInt(int64(k), 10)
		feeratesResult[asStr] = v
	}

	prices, err := s.pf.FetchPrice(USD, EUR, JPY, CNY, BRL)
	if err != nil {
		return stackerr.Wrap(err)
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
		return stackerr.Wrap(err)
	}
	return nil
}

func (s *server) routesplusHandler(
	w http.ResponseWriter, r *http.Request,
) error {
	log.Println("request on POST /router/routesplus")
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	err := r.ParseForm()
	if err != nil {
		return stackerr.Wrap(err)
	}
	params, err := decodeInRoutes([]byte(r.PostFormValue("params")))
	if err != nil {
		return stackerr.Wrap(err)
	}
	log.Printf("-> %+v", params)
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")

	routes, err := s.rf.FindRoutes(params.From, params.To, params.Sat*1000)
	if err != nil {
		log.Println("error getting routes, returning empty routes: ", err)
		routes = []PaymentRoute{}
	}

	result := []interface{}{
		"ok",
		serializeRoutes(routes),
	}
	log.Printf("<- %+v\n", result)
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		return stackerr.Wrap(err)
	}
	return nil
}

type PriceFetcher interface {
	FetchPrice(symbols ...Symbol) (map[Symbol]float64, error)
}

type FeerateFetcher interface {
	FetchFeerate(nBlockTarget ...int32) (map[int32]float64, error)
}

type RouteFinder interface {
	FindRoutes(
		fromPubkeys []string, toPubkey string, msat int64,
	) ([]PaymentRoute, error)
}

type inRoutes struct {
	Sat      int64    `json:"sat"`
	BadNodes []string `json:"badNodes"`
	BadChans []int64  `json:"badChans"`
	From     []string `json:"from"`
	To       string   `json:"to"`
}
