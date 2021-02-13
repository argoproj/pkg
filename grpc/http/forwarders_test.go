package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/argoproj/pkg/ticker"
)

type fakeTicker struct {
	c          chan time.Time
	resetCalls int
}

func (ft *fakeTicker) Stop() {
}

func (ft *fakeTicker) Reset(time.Duration) {
	ft.resetCalls++
}

func (ft *fakeTicker) C() <-chan time.Time {
	return ft.c
}

func (ft *fakeTicker) tick() {
	ft.c <- time.Now()
}

func newFakeTicker(time.Duration) *fakeTicker {
	return &fakeTicker{
		c:          make(chan time.Time, 1),
		resetCalls: 0,
	}
}

type testStruct struct {
	Metadata *testStruct `json:"metadata,omitempty"`
	Spec     *testStruct `json:"spec,omitempty"`
	Source   *testStruct `json:"source,omitempty"`
	Status   *testStruct `json:"status,omitempty"`

	Name    string `json:"name,omitempty"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message,omitempty"`
}

var (
	testVal = testStruct{
		Metadata: &testStruct{Name: "test"},
		Spec: &testStruct{
			Source: &testStruct{
				Path: "test_path",
			},
		},
		Status: &testStruct{
			Message: "Failed",
		},
	}
)

func TestMarshalerIncludeFields(t *testing.T) {
	m := messageMarshaler{fields: map[string]interface{}{
		"metadata.name": true,
		"spec.source":   true,
	}, exclude: false}

	out, err := m.Marshal(testVal)

	assert.Nil(t, err)
	assert.Equal(t, `{"metadata":{"name":"test"},"spec":{"source":{"path":"test_path"}}}`, string(out))
}

func TestMarshalerExcludeFields(t *testing.T) {
	m := messageMarshaler{fields: map[string]interface{}{
		"metadata.name": true,
	}, exclude: true}

	out, err := m.Marshal(testVal)

	assert.Nil(t, err)
	assert.Equal(t, `{"metadata":{},"spec":{"source":{"path":"test_path"}},"status":{"message":"Failed"}}`, string(out))
}

func TestMarshalerSSE(t *testing.T) {
	m := messageMarshaler{isSSE: true}

	out, err := m.Marshal(testVal)

	assert.Nil(t, err)
	assert.Equal(t, `data: {"metadata":{"name":"test"},"spec":{"source":{"path":"test_path"}},"status":{"message":"Failed"}}

`, string(out))
}

func Test_WithKeepalive(t *testing.T) {
	rr := httptest.NewRecorder()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ft := newFakeTicker(time.Second)
	wg := sync.WaitGroup{}

	var w http.ResponseWriter
	var recv recvFn

	w, recv = withKeepaliveAux(ctx, rr, func() (proto.Message, error) {
		_, _ = w.Write([]byte("data: 1\n"))
		return nil, nil
	}, &wg, func(d time.Duration) ticker.Ticker {
		return ft
	})

	wg.Add(1)
	ft.tick()
	wg.Wait()

	_, _ = recv()
	assert.Equal(t, 1, ft.resetCalls)
	_, _ = recv()
	assert.Equal(t, 2, ft.resetCalls)

	wg.Add(1)
	ft.tick()
	wg.Wait()

	assert.Equal(t, 2, ft.resetCalls)
	assert.Equal(t, ":\ndata: 1\ndata: 1\n:\n", rr.Body.String())
}
