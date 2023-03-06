package main

import (
	"os"

	d "github.com/alidevhere/go-fast-downloader"
)

func main() {
	//"http://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4"
	downloader, err := d.NewConcurrentDownloader(d.Options{
		Url:                 "https://test-videos.co.uk/vids/bigbuckbunny/mp4/h264/1080/Big_Buck_Bunny_1080_10s_5MB.mp4",
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
	os.Remove("test.mp4")
}
