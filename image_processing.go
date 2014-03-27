package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/rainycape/magick"
)

func downloadAndSaveOriginal(ic *ImageConfiguration) error {
	path := ic.OriginalImagePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		start := time.Now()

		remoteUrl := ic.RemoteImageUrl()
		resp, err := http.Get(remoteUrl)

		log.Printf("response code %d", resp.StatusCode)
		if err != nil || resp.StatusCode != 200 {
			log.Printf("Unable to download image: %s, status code: %d", remoteUrl, resp.StatusCode)
			log.Println(err)
			return fmt.Errorf("Unable to download image: %s, status code: %d", remoteUrl, resp.StatusCode)
		}
		defer resp.Body.Close()

		dir := filepath.Dir(path)
		os.MkdirAll(dir, 0700)

		out, err := os.Create(path)
		defer out.Close()
		if err != nil {
			log.Printf("Unable to create file: %s", path)
			log.Println(err)
			return fmt.Errorf("Unable to create file: %s", path)
		}

		io.Copy(out, resp.Body)
		log.Printf("Took %s to download image: %s", time.Since(start), path)
	}
	return nil
}

func createWithMagick(ic *ImageConfiguration) {
	start := time.Now()
	fullSizePath := ic.OriginalImagePath()
	im, err := magick.DecodeFile(fullSizePath)
	if err != nil {
		log.Panicln(err)
		return
	}
	defer im.Dispose()

	im2, err := im.CropResize(ic.width, ic.height, magick.FHamming, magick.CSCenter)
	if err != nil {
		log.Panicln(err)
		return
	}

	resizedPath := ic.ResizedImagePath()
	dir := filepath.Dir(resizedPath)
	os.MkdirAll(dir, 0700)

	out, err := os.Create(resizedPath)
	defer out.Close()

	info := ic.MagickInfo()
	err = im2.Encode(out, info)

	if err != nil {
		log.Panicln(err)
		return
	}
	elapsed := time.Since(start)
	log.Printf("Took %s to generate image: %s", elapsed, resizedPath)
}

func createImages(ic *ImageConfiguration) (string, error) {
	resizedPath := ic.ResizedImagePath()

	if _, err := os.Stat(resizedPath); os.IsNotExist(err) {
		err := downloadAndSaveOriginal(ic)
		log.Printf("what errors? %v", err)
		if err != nil {
			log.Printf("--something happened, skipping creation")
			return "", err
		}

		createWithMagick(ic)
	}

	return resizedPath, nil
}
