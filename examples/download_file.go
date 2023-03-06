package main

import (
	d "github.com/alidevhere/go-fast-downloader"
)

func main() {
	downloader, err := d.NewConcurrentDownloader(d.Options{
		Url:                 "http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4",
		OutputFileName:      "test.mp4",
		OutputFileDirectory: "./",
	})

	if err != nil {
		panic(err)
	}

	err = downloader.StartDownload()
	if err != nil {
		panic(err)
	}

}
