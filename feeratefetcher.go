package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"master.private/bstd.git/stackerr"
)

type feerateFetcher struct {
	c           *http.Client
	btcUrl      string
	btcUser     string
	btcPassword string
	feerates    map[int32]feerateItem
	mu          sync.Mutex
	count       int64
	maxCacheAge time.Duration
}

func NewFeerateFetcher(btcUrl, user, password string) *feerateFetcher {
	return &feerateFetcher{
		c:           &http.Client{Timeout: time.Second * 60},
		btcUrl:      btcUrl,
		btcUser:     user,
		btcPassword: password,
		mu:          sync.Mutex{},
		feerates:    map[int32]feerateItem{},
		maxCacheAge: time.Minute * 5,
	}

}

func (f *feerateFetcher) FetchFeerate(
	nBlockTarget ...int32,
) (map[int32]float64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	result := map[int32]float64{}
	var toFetch []int32
	expired := time.Now().Add(f.maxCacheAge * -1)
	for _, v := range nBlockTarget {
		fromState, ok := f.feerates[v]
		if ok && fromState.Time.After(expired) {
			result[v] = fromState.BtcPerKVByte
			continue
		}
		toFetch = append(toFetch, v)
	}

	err := f.fetchExtFeerates(toFetch...)
	if err != nil {
		return nil, stackerr.Wrap(err)
	}
	for _, v := range toFetch {
		result[v] = f.feerates[v].BtcPerKVByte
	}

	return result, nil
}

func (f *feerateFetcher) fetchExtFeerates(nBlockTarget ...int32) error {
	log.Println("fetching data from bitcoind")
	now := time.Now()
	for _, v := range nBlockTarget {
		params := struct {
			ConfTarget int32 `json:"conf_target"`
		}{v}
		result := struct {
			FeerateBtcKb float64         `json:"feerate"`
			Errors       json.RawMessage `json:"errors"`
			Blocks       int32           `json:"blocks"`
		}{}
		err := f.fromBtcRpc("estimatesmartfee", params, &result)
		if err != nil {
			return stackerr.Wrap(err)
		}

		f.feerates[v] = feerateItem{
			result.FeerateBtcKb, now,
		}
	}
	return nil
}

func (f *feerateFetcher) fromBtcRpc(
	method string, params, result interface{},
) error {
	defer func() { f.count++ }()

	rpcReq := struct {
		JsonRpc string      `json:"jsonrpc"`
		Id      int64       `json:"id"`
		Method  string      `json:"method"`
		Params  interface{} `json:"params"`
	}{"2.0", f.count, method, params}

	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(rpcReq)
	if err != nil {
		return stackerr.Wrap(err)
	}

	req, err := http.NewRequest("POST", f.btcUrl, buf)
	if err != nil {
		return stackerr.Wrap(err)
	}
	req.SetBasicAuth(f.btcUser, f.btcPassword)

	res, err := f.c.Do(req)
	if err != nil {
		return stackerr.Wrap(err)
	}
	defer res.Body.Close()
	buf.Reset()

	rpcRes := struct {
		JsonRpc string        `json:"jsonrpc"`
		Id      int64         `json:"id"`
		Result  interface{}   `json:"result"`
		Error   *jsonRpcError `json:"error"`
	}{Result: result}
	err = json.NewDecoder(res.Body).Decode(&rpcRes)
	if err != nil {
		return stackerr.Wrap(err)
	}
	if rpcRes.Error != nil {
		return stackerr.Wrap(rpcRes.Error)
	}

	return nil
}

type feerateItem struct {
	BtcPerKVByte float64
	Time         time.Time
}

type jsonRpcError struct {
	Code    int64           `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (e *jsonRpcError) Error() string {
	return fmt.Sprintf(
		"jsonrpc error. code: %d; message:%s, data: %s\n",
		e.Code,
		e.Message,
		e.Data,
	)
}
