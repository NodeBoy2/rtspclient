package main

import (
	"flag"
	"log"
	"net/http"
)

var (
	listen = flag.String("listen", ":8070", "listen address")
	dir    = flag.String("dir", ".", "directory to server")
)

func main() {
	flag.Parse()
	log.Printf("listening on %q", *listen)
	log.Fatal(http.ListenAndServe(*listen, http.FileServer(http.Dir(*dir))))
}
