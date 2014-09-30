package http

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Fetcher struct{}

func (f *Fetcher) Fetch(url string, destination string) error {
	if _, err := os.Stat(destination); os.IsNotExist(err) {
		start := time.Now()

		timeout := 5 * time.Second
		tr := &http.Transport{ResponseHeaderTimeout: timeout}
		client := &http.Client{Transport: tr}
		resp, err := client.Get(url)

		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("Unable to download image: %s, status code: %d", url, resp.StatusCode)
		}

		log.Printf("Downloaded from %s with code %d", url, resp.StatusCode)

		dir := filepath.Dir(destination)
		os.MkdirAll(dir, 0700)

		out, err := os.Create(destination)
		if err != nil {
			log.Printf("Unable to create file: %s", destination)
			log.Println(err)
			return fmt.Errorf("Unable to create file: %s", destination)
		}
		defer out.Close()

		io.Copy(out, resp.Body)
		log.Printf("Took %s to download image: %s", time.Since(start), destination)
	} else {
		log.Printf("Fetcher: image is already present on destination: %s", destination)
	}
	return nil
}
