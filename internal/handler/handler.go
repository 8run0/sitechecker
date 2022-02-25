package handler

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"
)

type LatencyRequestHandler struct {
	resChan  chan LatencyResponse
	doneChan chan struct{}
}

type LatencyRequest struct {
	site      string
	transport *LatencyTransport
	trace     *httptrace.ClientTrace
	client    http.Client
}

type LatencyResponse struct {
	site      string
	timeTaken time.Duration
}

func (h *LatencyRequestHandler) AsyncLatencyCheckSiteList(sites []string) {
	var wg sync.WaitGroup
	go func() {
		for _, site := range sites {
			wg.Add(1)
			latencyRequest := h.NewLatencyRequest(site)
			go func() {
				h.handle(&wg, latencyRequest)
			}()
		}
		wg.Wait()
		h.doneChan <- struct{}{}
	}()
	h.ListenForLatencyResponses()
}

func (h *LatencyRequestHandler) ListenForLatencyResponses() {
	for {
		select {
		case res := <-h.resChan:
			fmt.Printf("%s @ %s\n", res.site, res.timeTaken)
		case <-h.doneChan:
			fmt.Printf("Done")
			return
		}
	}
}

func (*LatencyRequestHandler) NewLatencyRequest(site string) *LatencyRequest {
	transport := NewTransport()
	trace := &httptrace.ClientTrace{
		ConnectStart: transport.ConnectStart,
		ConnectDone:  transport.ConnectDone,
	}
	client := &http.Client{Transport: transport}
	return &LatencyRequest{site: site, transport: transport, trace: trace, client: *client}
}

func (h *LatencyRequestHandler) handle(wg *sync.WaitGroup, latencyReq *LatencyRequest) {
	defer wg.Done()
	req, _ := http.NewRequest("GET", latencyReq.site, nil)
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), latencyReq.trace))
	_, err := latencyReq.client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	timeTaken := latencyReq.transport.TimeTaken()
	h.resChan <- LatencyResponse{site: latencyReq.site, timeTaken: timeTaken}
}

func NewLatencyRequestHandler() *LatencyRequestHandler {
	resChan := make(chan LatencyResponse, 1)
	doneChan := make(chan struct{}, 1)
	latencyRequestHandler := &LatencyRequestHandler{resChan: resChan, doneChan: doneChan}
	return latencyRequestHandler
}

type LatencyTransport struct {
	connStart  time.Time
	connEnd    time.Time
	traceError error
}

func NewTransport() *LatencyTransport {
	transport := &LatencyTransport{}
	return transport
}

func (t *LatencyTransport) TimeTaken() time.Duration {
	return t.connEnd.Sub(t.connStart)
}

func (t *LatencyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return http.DefaultTransport.RoundTrip(req)
}

func (t *LatencyTransport) ConnectStart(network, addr string) {
	if t.connStart.IsZero() {
		t.connStart = time.Now()
	}
}

func (t *LatencyTransport) ConnectDone(network, addr string, err error) {
	if err != nil {
		t.traceError = err
		return
	}
	t.connEnd = time.Now()
}

func NewTrace(transport *LatencyTransport) *httptrace.ClientTrace {
	trace := &httptrace.ClientTrace{
		ConnectStart: transport.ConnectStart,
		ConnectDone:  transport.ConnectDone,
	}
	return trace
}
