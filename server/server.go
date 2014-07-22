package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/wanelo/image-server/core"
	"github.com/wanelo/image-server/fetcher"
	"github.com/wanelo/image-server/info"
)

func InitializeRouter(sc *core.ServerConfiguration, port string) {
	log.Println("starting server on http://0.0.0.0:" + port)

	r := render.New(render.Options{
		IndentJSON: true,
	})

	router := mux.NewRouter()
	router.HandleFunc("/{namespace:[a-z0-9]+}", func(wr http.ResponseWriter, req *http.Request) {
		NewImageHandler(wr, req, sc, r)
	}).Methods("POST").Name("newImage")

	// n := negroni.New()
	n := negroni.Classic()
	n.UseHandler(router)

	n.Run(":" + port)
}

func NewImageHandler(w http.ResponseWriter, req *http.Request, sc *core.ServerConfiguration, r *render.Render) {
	qs := req.URL.Query()
	vars := mux.Vars(req)
	errorStr := ""

	source := qs.Get("source")
	namespace := vars["namespace"]

	f := fetcher.NewOriginalFetcher(sc.Adapters.Paths, sc.Adapters.Fetcher)
	fc := f.Channels
	err, imageDetails := f.Fetch(source, namespace)

	if err != nil {
		errorStr = fmt.Sprintf("%s", err)
	}

	hash := imageDetails.Hash

	go func() {
		select {
		case source := <-fc.DownloadComplete:
			destination := sc.Adapters.Paths.RemoteOriginalPath(namespace, hash)

			go sc.Adapters.Logger.OriginalDownloaded(source, destination)
			go func() {
				localInfoPath := sc.Adapters.Paths.LocalInfoPath(namespace, hash)
				remoteInfoPath := sc.Adapters.Paths.RemoteInfoPath(namespace, hash)

				info.SaveImageDetail(imageDetails, localInfoPath)
				sc.Adapters.Uploader.Upload(localInfoPath, remoteInfoPath)
			}()
			sc.Adapters.Uploader.Upload(source, destination)
		case <-fc.DownloadFailed:
			go sc.Adapters.Logger.OriginalDownloadFailed(source)
		case <-fc.SkippedDownload:
			go sc.Adapters.Logger.OriginalDownloadSkipped(source)
		}
	}()

	json := map[string]string{
		"error":  errorStr,
		"hash":   hash,
		"height": fmt.Sprintf("%v", imageDetails.Height),
		"width":  fmt.Sprintf("%v", imageDetails.Width),
	}

	r.JSON(w, http.StatusOK, json)
}
