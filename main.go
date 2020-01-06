package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/yunomu/httppipe/server"
)

var (
	bind = flag.String("bind", "localhost:8080", "HTTP bind")
)

func init() {
	flag.Parse()
	log.SetOutput(os.Stderr)
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", server.NewServer())

	log.Println("start")
	if err := http.ListenAndServe(*bind, mux); err != nil {
		log.Fatalln(err)
	}
}
