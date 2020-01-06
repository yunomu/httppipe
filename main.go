package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/yunomu/httppipe/handler"
)

var (
	bind = flag.String("bind", "localhost:8080", "HTTP bind")
)

func init() {
	flag.Parse()
	log.SetOutput(os.Stderr)
}

type logger struct{}

func (l *logger) Error(err error) {
	log.Println("Error", err)
}

func (l *logger) Info(msg string) {
	log.Println("Info", msg)
}

var _ handler.Logger = (*logger)(nil)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", handler.NewHandler(
		handler.SetLogger(&logger{}),
	))

	if err := http.ListenAndServe(*bind, mux); err != nil {
		log.Fatalln(err)
	}
}
