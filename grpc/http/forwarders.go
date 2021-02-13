package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/argoproj/pkg/ticker"
)

type recvFn func() (proto.Message, error)
type tickerFac func(time.Duration) ticker.Ticker

type messageMarshaler struct {
	fields  map[string]interface{}
	exclude bool
	isSSE   bool
}

func (m *messageMarshaler) Unmarshal(data []byte, v interface{}) error {
	return nil
}

func (m *messageMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return nil
}

func (m *messageMarshaler) NewEncoder(w io.Writer) runtime.Encoder {
	return nil
}

func (m *messageMarshaler) ContentType() string {
	if m.isSSE {
		return "text/event-stream"
	} else {
		return "application/json"
	}
}

func (m *messageMarshaler) Marshal(v interface{}) ([]byte, error) {
	dataBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(m.fields) > 0 {
		if _, ok := v.([]interface{}); ok {
			data := make([]interface{}, 0)
			err = json.Unmarshal(dataBytes, &data)
			if err != nil {
				return nil, err
			}
			for i := range data {
				m.processItem([]string{}, data[i])
			}
			dataBytes, err = json.Marshal(data)
			if err != nil {
				return nil, err
			}
		} else {
			data := make(map[string]interface{})
			err = json.Unmarshal(dataBytes, &data)
			if err != nil {
				return nil, err
			}
			m.processItem([]string{}, data)
			dataBytes, err = json.Marshal(data)
			if err != nil {
				return nil, err
			}
		}
	}
	if m.isSSE {
		dataBytes = []byte(fmt.Sprintf("data: %s \n\n", string(dataBytes)))
	}
	return dataBytes, nil
}

func (m *messageMarshaler) processItem(path []string, item interface{}) {
	if mapItem, ok := item.(map[string]interface{}); ok {
		for k, v := range mapItem {
			fieldPath := strings.Join(append(path, k), ".")
			_, pathIn := m.fields[fieldPath]
			parentPathIn := pathIn
			if !parentPathIn {
				for k := range m.fields {
					if strings.HasPrefix(k, fieldPath) {
						parentPathIn = true
						break
					}
				}
			}
			keep := m.exclude && !pathIn || !m.exclude && parentPathIn

			if keep {
				if !pathIn {
					m.processItem(append(path, k), v)
				}
			} else {
				delete(mapItem, k)
			}
		}
	} else if arrayItem, ok := item.([]interface{}); ok {
		for i := range arrayItem {
			m.processItem(path, arrayItem[i])
		}
	}
}

func newMarshaler(req *http.Request, isSSE bool) runtime.Marshaler {
	fieldsQuery := req.URL.Query().Get("fields")
	fields := make(map[string]interface{})
	exclude := false
	if fieldsQuery != "" {
		if strings.HasPrefix(fieldsQuery, "-") {
			fieldsQuery = fieldsQuery[1:]
			exclude = true
		}
		for _, field := range strings.Split(fieldsQuery, ",") {
			fields[field] = true
		}
	}
	return &messageMarshaler{isSSE: isSSE, fields: fields, exclude: exclude}
}

type StreamForwarderFunc func(
	ctx context.Context,
	mux *runtime.ServeMux,
	marshaler runtime.Marshaler,
	w http.ResponseWriter,
	req *http.Request,
	recv func() (proto.Message, error),
	opts ...func(context.Context, http.ResponseWriter, proto.Message) error,
)

func writeKeepalive(w http.ResponseWriter, mut *sync.Mutex) {
	mut.Lock()
	defer mut.Unlock()

	// Per https://html.spec.whatwg.org/multipage/server-sent-events.html#event-stream-interpretation,
	// lines that start with a `:` must be ignored by the client.
	_, err := w.Write([]byte(":\n"))

	if err != nil {
		log.Warnf("failed to write http keepalive response: %v", err)
	} else if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func keepalive(ctx context.Context, w http.ResponseWriter, mut *sync.Mutex, wg *sync.WaitGroup, t ticker.Ticker) {
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C():
			writeKeepalive(w, mut)

			// The waitgroup is purely intended for unit tests and is always nil in production use cases
			if wg != nil {
				wg.Done()
			}
		}
	}
}

func withKeepalive(ctx context.Context, w http.ResponseWriter, recv recvFn) (http.ResponseWriter, recvFn) {
	return withKeepaliveAux(ctx, w, recv, nil, ticker.NewTicker)
}

func withKeepaliveAux(ctx context.Context, w http.ResponseWriter, recv recvFn, wg *sync.WaitGroup, tf tickerFac) (http.ResponseWriter, recvFn) {
	keepaliveInterval := time.Duration(time.Second * 15)
	keepaliveTicker := tf(keepaliveInterval)
	oldRecv := recv
	mut := sync.Mutex{}
	oldw := w

	w = httpsnoop.Wrap(oldw, httpsnoop.Hooks{
		Write: func(next httpsnoop.WriteFunc) httpsnoop.WriteFunc {
			return func(p []byte) (int, error) {
				mut.Lock()
				defer mut.Unlock()
				return next(p)
			}
		},
	})

	recv = func() (proto.Message, error) {
		result, err := oldRecv()

		if ctx.Err() == nil {
			keepaliveTicker.Reset(keepaliveInterval)
		}

		return result, err
	}

	go keepalive(ctx, oldw, &mut, wg, keepaliveTicker)

	return w, recv
}

func NewStreamForwarder(messageKey func(proto.Message) (string, error)) StreamForwarderFunc {
	return func(
		ctx context.Context,
		mux *runtime.ServeMux,
		marshaler runtime.Marshaler,
		w http.ResponseWriter,
		req *http.Request,
		recv func() (proto.Message, error),
		opts ...func(context.Context, http.ResponseWriter, proto.Message) error,
	) {
		isSSE := req.Header.Get("Accept") == "text/event-stream"

		if isSSE {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Transfer-Encoding", "chunked")
			w.Header().Set("X-Content-Type-Options", "nosniff")
		}
		dataByKey := make(map[string][]byte)
		m := newMarshaler(req, isSSE)
		if messageKey != nil {
			oldRecv := recv
			recv = func() (proto.Message, error) {
				for {
					if ctx.Err() != nil {
						return nil, ctx.Err()
					}

					msg, err := oldRecv()
					if err != nil {
						return nil, err
					}
					key, err := messageKey(msg)
					if err != nil {
						return nil, err
					}
					data, err := m.Marshal(map[string]proto.Message{"result": msg})
					if err != nil {
						return nil, err
					}
					prevData, ok := dataByKey[key]
					if !ok || string(prevData) != string(data) {
						dataByKey[key] = data
						return msg, nil
					}
				}
			}
		}

		if isSSE && os.Getenv("DISABLE_SSE_KEEPALIVE") == "" {
			w, recv = withKeepalive(ctx, w, recv)
		}

		runtime.ForwardResponseStream(ctx, mux, m, w, req, recv, opts...)
	}
}

var (
	// UnaryForwarder serializes protobuf message to JSON and removes fields using query parameter `fields`.
	// The `fields` parameter example:
	// fields=items.metadata.name,items.spec - response should include only items.metadata.name and items.spec fields
	// fields=-items.metadata.name - response should include all fields except items.metadata.name
	UnaryForwarder = func(
		ctx context.Context,
		mux *runtime.ServeMux,
		marshaler runtime.Marshaler,
		w http.ResponseWriter,
		req *http.Request,
		resp proto.Message,
		opts ...func(context.Context, http.ResponseWriter, proto.Message) error,
	) {
		runtime.ForwardResponseMessage(ctx, mux, newMarshaler(req, false), w, req, resp, opts...)
	}

	// StreamForwarder serializes protobuf message to JSON and removes fields using query parameter `fields`
	StreamForwarder = NewStreamForwarder(nil)
)
