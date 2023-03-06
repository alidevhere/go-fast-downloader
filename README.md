
# Go Fast Downloader
   This is concurrent downloader is for educational purposes. Please read Medium article to learn concurrency in go by implementing concurrent downloader.
   

# Add package
```
go get github.com/alidevhere/go-fast-downloader

```


## Example:
Check out examples directory or this example below:

```

import "github.com/alidevhere/go-fast-downloader"


func main() {
	url := "https://test-videos.co.uk/vids/bigbuckbunny/mp4/h264/1080/Big_Buck_Bunny_1080_10s_5MB.mp4"
	options := Options{
		Url:                 url,
		ChunkSizeInBytes:    1024 * 1024 * 5,
		Concurrency:         5,
		Retries:             3,
		OutputFileDirectory: ".",
		OutputFileName:      "output.mp4",
	}

	d, err := NewConcurrentDownloader(options)
	if err != nil {
		println(err.Error())
	}

	err = d.StartDownload()
	if err != nil {
		println(err.Error())
	}
	println(d.DownloadTime().Seconds())
}

```