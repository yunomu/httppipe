package server

import (
	"errors"
	"io"
	"net/http"
)

type Logger interface {
	Error(error)
}

type NopLogger struct{}

func (l *NopLogger) Error(_ error) {
}

type Server struct {
	logger  Logger
	streams map[string]chan io.Reader
}

func NewServer() *Server {
	return &Server{
		logger:  &NopLogger{},
		streams: make(map[string]chan io.Reader),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ch, ok := s.streams[r.URL.String()]
		if !ok {
			ch = make(chan io.Reader)
			s.streams[r.URL.String()] = ch
		}

		select {
		case in := <-ch:
			if _, err := io.Copy(w, in); err != nil {
				s.logger.Error(err)
			}
		case <-r.Context().Done():
			s.logger.Error(errors.New("request timeout"))
			w.WriteHeader(http.StatusRequestTimeout)
		}
	case http.MethodPut:
		if _, ok := s.streams[r.URL.String()]; ok {
			s.logger.Error(errors.New("not found"))
			w.WriteHeader(http.StatusNotFound)
		}

		defer r.Body.Close()

		ch := make(chan io.Reader)
		s.streams[r.URL.String()] = ch

		select {
		case ch <- r.Body:
		case <-r.Context().Done():
			s.logger.Error(errors.New("request timeout"))
			w.WriteHeader(http.StatusRequestTimeout)
		}
	default:
		s.logger.Error(errors.New("method not allowed"))
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
