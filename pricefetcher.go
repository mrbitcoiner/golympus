package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"master.private/bstd.git/stackerr"
)

type Symbol = string

const (
	USD Symbol = "usd"
	EUR Symbol = "eur"
	JPY Symbol = "jpy"
	CNY Symbol = "cny"
	BRL Symbol = "brl"
)

type priceFetcher struct {
	state map[Symbol]priceData
	stateMu sync.Mutex
	maxLife time.Duration
	c *http.Client
	fetchCount int64
}

func NewPriceFetcher() *priceFetcher {
	return &priceFetcher{
		state: map[Symbol]priceData{},
		stateMu: sync.Mutex{},
		maxLife: time.Minute * 5,
		c: &http.Client{Timeout: time.Second * 60},
		fetchCount: 0,
	}
}

type priceData struct {
	price float64
	fetchTime time.Time
}

func (p *priceFetcher) Fetch(symbols ...Symbol) (map[Symbol]float64, error) {
	if len(symbols) == 0 {
		return nil, nil
	}
	p.stateMu.Lock()
	defer p.stateMu.Unlock()
	if len(p.state) == 0 {
		err := p.extFetch(symbols...)
		if err != nil {
			return nil, stackerr.Wrap(err)
		}
	} 
	var missingSymbols []Symbol
	result := map[Symbol]float64{}
	now := time.Now()
	expired := now.Add(p.maxLife * -1)

	for _, v := range symbols {
		fromState, ok :=  p.state[v]
		if ok && fromState.fetchTime.After(expired) {
			result[v] = fromState.price
			continue
		}
		missingSymbols = append(missingSymbols, v)
	}
	if len(missingSymbols) == 0 {
		return result, nil
	}
	
	// only when there are missing or expired symbols
	err := p.extFetch(missingSymbols...)
	if err != nil {
		return nil, stackerr.Wrap(err)
	}
	for _, v := range missingSymbols {
		fromState, ok := p.state[v]
		if !ok {
			result[v] = 0
			continue
		}
		result[v] = fromState.price
	}
	return result, nil
}

func (p *priceFetcher) extFetch(sym ...Symbol) error {
	log.Println("fetching ext symbols")
	p.fetchCount++
	res := struct {
		BitcoinVs map[Symbol]float64 `json:"bitcoin"`
	}{}
	urlBuilder := strings.Builder{}
	urlBuilder.Write([]byte("https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies="))
	for _, v := range sym {
		urlBuilder.Write([]byte(v))
		urlBuilder.Write([]byte(","))
	}
	url := urlBuilder.String()
	response, err := p.c.Get(url)
	if err != nil {
		return stackerr.Wrap(err)
	}
	if s := response.StatusCode; s != 200 {
		return stackerr.Wrap(
			fmt.Errorf("invalid status calling coingecko api: %d", s),
		)
	}
	err = json.NewDecoder(response.Body).Decode(&res)
	if err != nil {
		return stackerr.Wrap(err)
	}
	now := time.Now()
	for k, v := range res.BitcoinVs {
		p.state[k] = priceData{
			price: v,
			fetchTime: now,
		}
	}
	for _, v := range sym {
		_, ok := p.state[v]
		if ok {
			continue
		}
		log.Printf("failure fetching symbol: %s\n", "btc"+v)
	}
	return nil
} 


