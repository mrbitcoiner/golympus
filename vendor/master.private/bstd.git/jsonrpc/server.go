package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"master.private/bstd.git/log"
	"master.private/bstd.git/stackerr"
)

type Server struct {
	log      *log.Logger
	handlers map[string]Handler
	pool     sync.Pool
}

func NewServer(log *log.Logger) *Server {
	return &Server{
		log:      log,
		handlers: map[string]Handler{},
		pool: sync.Pool{
			New: func() interface{} {
				txBuf := &bytes.Buffer{}
				return &SocketContext{
					rxBuf: []byte{},
					enc:   json.NewEncoder(txBuf),
					txBuf: txBuf,
				}
			},
		},
	}
}

func (s *Server) AddHandler(method string, hdl Handler) {
	s.handlers[method] = hdl
}

func (s *Server) ListenAndServe(
	ctx context.Context, network string, address string,
) error {
	var (
		wg          sync.WaitGroup
		allowAccept = make(chan struct{}, 1)
	)

	listener, err := net.Listen(network, address)
	if err != nil {
		return err
	}
	defer listener.Close()
	s.log.Infof("jsonrpc listening on %s://%s", network, address)

	allowAccept <- struct{}{}
	for {
		select {
		case <-ctx.Done():

			s.log.Info("jsonrpc: shutting down on context cancel")
			listener.Close()
			goto end

		case <-allowAccept:

			wg.Add(1)
			go func() {
				defer wg.Done()
				sock, err := listener.Accept()
				select {
				case allowAccept <- struct{}{}:
				default:
				}
				if errors.Is(err, net.ErrClosed) {
					return
				}
				if err != nil {
					s.log.Err(stackerr.Wrap(err))
					return
				}
				s.HandleSocket(ctx, sock)
			}()

		}
	}

end:
	wg.Wait()
	s.log.Info("jsonrpc: shutdown completed")
	return nil
}

func (s *Server) HandleSocket(ctx context.Context, sock io.ReadWriteCloser) {
	var (
		err error
	)
	defer sock.Close()
	sockCtx := s.pool.Get().(*SocketContext)
	defer s.pool.Put(sockCtx)
	s.log.Debug("jsonrpc: new client")

	for {
		select {
		case <-ctx.Done():
			s.log.Info("context cancelled, stop handling client socket")
			goto end
		default:
		}
		err = s.handle(sockCtx, sock)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			s.log.Err(stackerr.Wrap(err))
			break
		}
	}

end:
	s.log.Debug("jsonrpc: client EOF")
}

func (s *Server) handle(
	ctx *SocketContext, transport io.ReadWriter,
) (err error) {
	var (
		n   int
		req Request
	)
	defer func() {
		var rpcErr RpcError
		if err == nil {
			return
		}
		if !errors.As(err, &rpcErr) {
			return
		}
		res := Response{JsonRpc: "2.0", Id: req.Id, Error: &rpcErr}
		err1 := s.send(ctx, res, transport)
		if err1 != nil {
			err = errors.Join(err, err1)
		}
	}()

	ctx.rxBuf = ctx.rxBuf[:0]
	n, err = ReadTillDelimiter(&ctx.rxBuf, transport, '\n')
	if errors.Is(err, io.EOF) {
		return io.EOF
	} else if n == 0 {
		return ErrRpcInvalidRequest
	} else if err != nil {
		return stackerr.Wrap(err)
	}
	s.log.Debug("jsonrpc: request incoming")

	req.Raw = ctx.rxBuf
	err = json.Unmarshal(ctx.rxBuf, &req)
	if err != nil {
		return ErrRpcParse
	}
	if req.JsonRpc != "2.0" {
		return ErrRpcInvalidRequest
	}

	res := Response{
		Id:      req.Id,
		JsonRpc: "2.0",
	}
	hdl, ok := s.handlers[req.Method]
	if !ok {
		return ErrRpcMethodNotFound
	}

	err = hdl(&res, &req)
	if err != nil {
		return stackerr.Wrap(err)
	}

	err = s.send(ctx, res, transport)
	if err != nil {
		return stackerr.Wrap(err)
	}
	return nil
}

func (s *Server) send(ctx *SocketContext, res Response, sock io.Writer) error {
	ctx.txBuf.Reset()
	err := ctx.enc.Encode(res)
	if err != nil {
		return stackerr.Wrap(err)
	}
	buf := [1]byte{'\n'}
	bytes := ctx.txBuf.Bytes()
	if len(bytes) > 0 && bytes[len(bytes)-1] != '\n' {
		ctx.txBuf.Write(buf[:])
	}
	_, err = ctx.txBuf.WriteTo(sock)
	if err != nil {
		return stackerr.Wrap(err)
	}
	return nil
}

type Handler func(ResultSetter, ParamParser) error

type ParamParser interface {
	Parse(params interface{}) error
}

type ResultSetter interface {
	SetResult(r interface{})
}

type SocketContext struct {
	rxBuf []byte
	enc   *json.Encoder
	txBuf *bytes.Buffer
}

type Request struct {
	Raw     []byte
	JsonRpc string          `json:"jsonrpc"`
	Id      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

func (r *Request) Parse(params interface{}) error {
	return json.Unmarshal(r.Params, params)
}

type Response struct {
	JsonRpc string      `json:"jsonrpc"`
	Id      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RpcError   `json:"error,omitempty"`
}

func (r *Response) SetResult(result interface{}) {
	r.Result = result
}

var (
	ErrRpcParse          = RpcError{Code: -32700, Message: "Parse error"}
	ErrRpcInvalidRequest = RpcError{Code: -32600, Message: "Invalid Request"}
	ErrRpcMethodNotFound = RpcError{Code: -32601, Message: "Method not found"}
	ErrRpcInvalidParams  = RpcError{Code: -32602, Message: "Invalid params"}
	ErrRpcInternalError  = RpcError{Code: -32603, Message: "Internal error"}
)

type RpcError struct {
	Code    int64       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (r RpcError) Error() string {
	return fmt.Sprintf("jsonrpc error; code: %d; message: %s", r.Code, r.Message)
}

func ReadTillDelimiter(
	output *[]byte, source io.Reader, delim byte,
) (int, error) {
	var (
		nTotal = 0
		buf    [1]byte
		out    = *output
		n      int
		err    error
		stop   bool
	)
	defer func() { *output = out }()
	for !stop {
		n, err = source.Read(buf[:])
		if err != nil || n == 0 {
			stop = true
		} else if n > 0 && buf[0] == delim {
			return nTotal, nil
		}
		nTotal += n
		out = append(out, buf[0])
	}
	return nTotal, err
}
