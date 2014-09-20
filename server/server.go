package server

import (
	"log"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/wanelo/image-server/core"
)

func InitializeRouter(sc *core.ServerConfiguration, listen string, port string) {
	log.Printf("starting server on http://%s:%s", listen, port)

	router := mux.NewRouter()
	router.HandleFunc("/{namespace:[a-z0-9]+}", func(wr http.ResponseWriter, req *http.Request) {
		NewImageHandler(wr, req, sc)
	}).Methods("POST").Name("newImage")

	router.HandleFunc("/{namespace:[a-z0-9]+}/{id1:[a-f0-9]{3}}/{id2:[a-f0-9]{3}}/{id3:[a-f0-9]{3}}/{id4:[a-f0-9]{23}}/{filename}", func(wr http.ResponseWriter, req *http.Request) {
		ResizeHandler(wr, req, sc)
	}).Methods("GET").Name("resizeImage")

	router.HandleFunc("/{namespace}/batch", func(wr http.ResponseWriter, req *http.Request) {
		CreateBatchHandler(wr, req, sc)
	}).Methods("POST").Name("createBatch")

	router.HandleFunc("/{namespace}/batch/{uuid:[a-f0-9-]{36}}", func(wr http.ResponseWriter, req *http.Request) {
		BatchHandler(wr, req, sc)
	}).Methods("GET").Name("batch")

	router.HandleFunc("/status_check", StatusHandler)

	n := negroni.Classic()
	n.UseHandler(router)

	n.Run(listen + ":" + port)
}
