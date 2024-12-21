package jsonrpc

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"

	"master.private/bstd.git/stackerr"
)

type Client struct {
	counter   int64
	mu        sync.Mutex
	transport io.ReadWriteCloser
	txBuf     *bytes.Buffer
	enc       *json.Encoder
	rxBuf     []byte
}

func NewClient(transport io.ReadWriteCloser) *Client {
	txBuf := &bytes.Buffer{}
	return &Client{
		transport: transport,
		txBuf:     txBuf,
		enc:       json.NewEncoder(txBuf),
		rxBuf:     []byte{},
	}
}

func (c *Client) Close() error {
	return c.transport.Close()
}

func (c *Client) Call(
	method string, params interface{}, result interface{},
) error {
	var (
		count int64
		err   error
		endl  = [1]byte{'\n'}
		n int
	)

	c.mu.Lock()
	defer c.mu.Unlock()

	count = c.counter
	c.counter++

	request := ClientRequest{
		JsonRpc: "2.0",
		Id:      count,
		Method:  method,
		Params:  params,
	}

	c.txBuf.Reset()
	err = c.enc.Encode(&request)
	if err != nil {
		return stackerr.Wrap(err)
	}

	if bytes := c.txBuf.Bytes(); len(bytes) > 0 && bytes[len(bytes)-1] != '\n' {
		c.txBuf.Write(endl[:])
	}

	_, err = c.txBuf.WriteTo(c.transport)
	if err != nil {
		return stackerr.Wrap(err)
	}

	response := Response{
		Id:     int64(0),
		Result: result,
	}

	retry:
	c.rxBuf = c.rxBuf[:0]
	n, err = ReadTillDelimiter(&c.rxBuf, c.transport, '\n')
	if err != nil {
		return stackerr.Wrap(err)
	}
	if n == 0 {
		goto retry
	}

	err = json.Unmarshal(c.rxBuf, &response)
	if err != nil {
		return stackerr.Wrap(err)
	}

	if response.Error != nil {
		return response.Error
	}

	return nil
}

type ClientRequest struct {
	JsonRpc string      `json:"jsonrpc"`
	Id      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}
