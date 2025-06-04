package sse

import (
	"net/http"

	"github.com/r3labs/sse/v2"
)

type R3LabsServer struct {
	srv *sse.Server
}

func New() *R3LabsServer {
	return &R3LabsServer{srv: sse.New()}
}

func (s *R3LabsServer) ServeHTTP(w http.ResponseWriter, r *http.Request, channel string) {
	q := r.URL.Query()
	q.Set("stream", channel)
	r.URL.RawQuery = q.Encode()
	s.srv.ServeHTTP(w, r)
}

func (s *R3LabsServer) Publish(channel string, data []byte) error {
	s.srv.Publish(channel, &sse.Event{Data: data})
	return nil
}
