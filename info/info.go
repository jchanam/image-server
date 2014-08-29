package info

import (
	"crypto/md5"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"os"
)

type Info struct {
	Path string
}

func (i Info) FileHash() (hash string, err error) {
	var contents []byte
	if contents, err = ioutil.ReadFile(i.Path); err == nil {
		hash := fmt.Sprintf("%x", md5.Sum(contents))
		return hash, nil
	}
	return hash, err
}

// ImageDetails extracts file hash, height, and width when providing a image path
// it returns an ImageDetails object
func (i Info) ImageDetails() (*ImageDetails, error) {
	if reader, err := os.Open(i.Path); err == nil {
		defer reader.Close()
		im, _, err := image.DecodeConfig(reader)
		if err != nil {
			return nil, err
		}
		hash, err := i.FileHash()
		if err != nil {
			return nil, err
		}
		details := &ImageDetails{
			Hash:   hash,
			Height: im.Height,
			Width:  im.Width,
		}
		return details, nil
	} else {
		return nil, err
	}
}
