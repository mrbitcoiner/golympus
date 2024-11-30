package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func must(err error) {
	if err == nil {
		return
	}
	panic(err)
}

func main() {
	srv := newServer()
	http.HandleFunc("POST /rates/get", srv.ratesHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("request on /")
	})
	//http.ListenAndServeTLS("0.0.0.0:8085", "data/certs/server.crt", "data/certs/server.key", nil)
	http.ListenAndServe("0.0.0.0:8085", nil)
}

type server struct {}

func newServer() *server {
	return &server {}
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
	res := []interface{}{
		"ok",
		[]map[string]float64{
			{
				"3": 0.0,
				"6": 0.0,
			},
			{
				"usd": 100000.05000,
				"eur": 101000.10000,
				"jpy": 14436600.000,
				"cny": 50000.10000,
				"brl": 580000.20000,
			},
		},
	}
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Println(err)
	}
}
