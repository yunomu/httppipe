package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Logger interface {
	Error(error)
	Info(string)
}

type NopLogger struct{}

func (l *NopLogger) Error(_ error) {
}

func (l *NopLogger) Info(_ string) {
}

type stream struct {
	in   chan io.Reader
	done chan struct{}
}

func newStream() *stream {
	return &stream{
		in:   make(chan io.Reader),
		done: make(chan struct{}),
	}
}

type Handler struct {
	logger Logger

	streams map[string]*stream
}

type HandlerOption func(*Handler)

func SetLogger(l Logger) HandlerOption {
	return func(s *Handler) {
		s.logger = l
	}
}

func NewHandler(options ...HandlerOption) *Handler {
	s := &Handler{
		logger:  &NopLogger{},
		streams: make(map[string]*stream),
	}
	for _, f := range options {
		f(s)
	}

	return s
}

func (s *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.logger.Info(fmt.Sprintf("%s %s", r.Method, r.URL.String()))

	switch r.Method {
	case http.MethodGet:
		ch := s.streams[r.URL.String()]
		if ch == nil {
			ch = newStream()
			s.streams[r.URL.String()] = ch
		}

		select {
		case in := <-ch.in:
			if _, err := io.Copy(w, in); err != nil {
				s.logger.Error(err)
			}
			select {
			case ch.done <- struct{}{}:
				close(ch.done)
			case <-r.Context().Done():
				s.logger.Error(errors.New("ack timeout"))
				w.WriteHeader(http.StatusRequestTimeout)
			}
		case <-r.Context().Done():
			s.logger.Error(errors.New("request timeout"))
			w.WriteHeader(http.StatusRequestTimeout)
		}

		s.streams[r.URL.String()] = nil
	case http.MethodPut:
		ch := s.streams[r.URL.String()]
		if ch == nil {
			ch = newStream()
		}
		defer r.Body.Close()

		s.streams[r.URL.String()] = ch

		select {
		case ch.in <- r.Body:
			close(ch.in)
			select {
			case <-ch.done:
				return
			case <-r.Context().Done():
				s.logger.Error(errors.New("send timeout"))
				w.WriteHeader(http.StatusRequestTimeout)
			}
		case <-r.Context().Done():
			s.logger.Error(errors.New("request timeout"))
			w.WriteHeader(http.StatusRequestTimeout)
		}

		s.streams[r.URL.String()] = nil
	default:
		s.logger.Error(errors.New("method not allowed"))
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
