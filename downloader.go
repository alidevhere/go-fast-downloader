package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Options struct {
	// The number of bytes to download in each request.
	// If not set, the default is 1KB or 1024 Bytes.
	ChunkSizeInBytes int
	// The number of concurrent requests to make.
	// If not set, the default is 5.
	Concurrency int
	// The number of retries to make for each request.
	// If not set, the default is 3.
	Retries int
	// Url to download from.
	Url string
	// Output file path.
	outputFilePath string
}

type ConcurrentDownloader interface {
	StartDownload() error
}

type downloader struct {
	outputFilePath string
	firstChunkID   int
	lastChunkID    int
	client         *http.Client
	options        Options
	chunkRanges    []chunkRange
	wg             *sync.WaitGroup
	inputChan      chan chunkRange
	outputChan     chan dataChunk
	stopChan       chan struct{}
}

type chunkRange struct {
	chunkId int
	header  string
}

type dataChunk struct {
	chunkId      int
	data         []byte
	downloadTime int64 //ms
	retries      int
}

type info struct {
	Offset int64
	Length int
}

func (d *downloader) worker(inputChan <-chan chunkRange, outputChan chan<- dataChunk, stopChan <-chan struct{}, URL string) {
	defer d.wg.Done()

	for {
		select {
		case <-stopChan:
			fmt.Println("Stopping worker")
			return
		case in := <-inputChan:
			fmt.Println("Downloading chunk", in.chunkId)
			for i := 0; i < d.options.Retries; i++ {
				data, err := d.downloadChunk(in)
				if err != nil {
					fmt.Println("Error downloading chunk", in.chunkId, "retrying...")
					continue
				}
				data.retries = i
				outputChan <- data
				fmt.Println("Downloading chunk complete ", in.chunkId)
				break
			}

		}
	}
}

func newOffSet(f *os.File) int64 {
	fs, _ := f.Stat()
	s := fs.Size()
	fmt.Println("File size", s)
	return s
}

func (d *downloader) merger() error {
	defer d.wg.Done()
	tmpFileName := "temp.mp4"
	f, err := os.OpenFile(tmpFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	// defer os.Remove(f.Name())

	m := make(map[int]info)
	downloadedChunksMapping := make(map[int]bool)

	for {
		select {

		case data := <-d.outputChan:
			if _, ok := downloadedChunksMapping[data.chunkId]; !ok {
				downloadedChunksMapping[data.chunkId] = true
				fmt.Printf("Downloaded chunk %d in %d ms with %d retries", data.chunkId, data.downloadTime, data.retries)
			}

			offset := newOffSet(f)
			m[data.chunkId] = info{Offset: offset, Length: len(data.data)}
			_, err := f.Write(data.data)
			if err != nil {
				println(err.Error())
			}

			if len(downloadedChunksMapping) == len(d.chunkRanges) {
				// of, _ := os.Create(d.outputFilePath)
				of, e := os.OpenFile(d.outputFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
				if e != nil {
					println(e.Error())
				}

				defer of.Close()
				f.Close()
				sortOutputFiles(m, d.firstChunkID, d.lastChunkID, tmpFileName, of)
				for i := 0; i < d.options.Concurrency; i++ {
					d.stopChan <- struct{}{}
				}
				return nil
			}
		}

	}

	// return nil
}

func sortOutputFiles(m map[int]info, firstID, lastID int, inputTmpFileName string, output *os.File) {

	input, e := os.Open(inputTmpFileName)
	if e != nil {
		fmt.Println("Error opening input file", e)
		return
	}

	for i := firstID; i <= lastID; i++ {

		info, ok := m[i]
		if !ok {
			fmt.Println("Missing chunk", i)
			continue
		}
		data := make([]byte, info.Length)
		readB, e := input.ReadAt(data, info.Offset)
		if e != nil {
			fmt.Println("Error reading chunk from", i, e)
			continue
		}
		writeB, _ := output.Write(data)
		fmt.Println("Writing chunk", i, "expected length", info.Length, "actual read", readB, "actual write", writeB)
	}

}

func (d *downloader) downloadChunk(info chunkRange) (dataChunk, error) {
	var data dataChunk
	r, err := http.NewRequest("GET", d.options.Url, nil)
	if err != nil {
		return data, err
	}

	r.Header.Add("Range", info.header)
	startTime := time.Now()
	res, err := d.client.Do(r)
	data.downloadTime = time.Since(startTime).Milliseconds()
	if err != nil {
		return data, err
	}

	data.data, err = ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	data.chunkId = info.chunkId
	if err != nil {
		return data, err
	}

	return data, nil
}

func NewConcurrentDownloader(options Options) (ConcurrentDownloader, error) {
	if options.ChunkSizeInBytes == 0 {
		options.ChunkSizeInBytes = 1024
	}

	if options.Concurrency == 0 {
		options.Concurrency = 5
	}

	if options.Retries == 0 {
		options.Retries = 3
	}

	if options.Url == "" {
		return nil, errors.New("Url is required")
	}

	if options.outputFilePath == "" {
		return nil, errors.New("outputFilePath is required")
	}

	return &downloader{options: options, client: &http.Client{
				// Timeout: time.Second * 10,
	}, outputFilePath: options.outputFilePath}, nil

}

func (d *downloader) StartDownload() error {

	err := d.calculateChunkRanges()
	if err != nil {
		return err
	}

	d.inputChan = make(chan chunkRange, len(d.chunkRanges))
	d.outputChan = make(chan dataChunk, len(d.chunkRanges))
	d.stopChan = make(chan struct{})
	d.wg = &sync.WaitGroup{}
	d.wg.Add(d.options.Concurrency + 1)
	for i := 0; i < d.options.Concurrency; i++ {
		go d.worker(d.inputChan, d.outputChan, d.stopChan, d.options.Url)
	}

	go d.merger()

	for _, chunk := range d.chunkRanges {
		d.inputChan <- chunk
	}

	d.wg.Wait()

	return nil

}

func (d *downloader) requestHEAD() (int, error) {

	fmt.Println("Requesting HEAD")

	r, err := http.NewRequest("HEAD", d.options.Url, nil)
	if err != nil {
		return 0, err
	}

	res, err := d.client.Do(r)
	if err != nil {
		return 0, err
	}

	fmt.Printf("Response: %+v", res)

	if res.StatusCode < 200 || res.StatusCode > 206 {
		return 0, errors.New(fmt.Sprintf("Status Code (%d): Invalid header received", res.StatusCode))
	}

	cl := res.Header.Get("content-length")
	length, err := strconv.Atoi(cl)
	if err != nil {
		return length, err
	}

	name := res.Header.Get("Content-Disposition")
	println("Content-Disposition", name)

	//it will be 'none' if ranges not supported
	if res.Header.Get("Accept-Ranges") == "none" {
		fmt.Println("Ranges not supported by server")
		return length, errors.New("Ranges not supported by server")
	}

	fmt.Println("File length: ", length, " bytes")

	return length, nil
}

func (d *downloader) calculateChunkRanges() error {

	length, err := d.requestHEAD()
	if err != nil {
		return err
	}

	chunk_id := 1
	for lowerIndex := 0; lowerIndex < length; {

		uperIndex := lowerIndex + d.options.ChunkSizeInBytes - 1

		if uperIndex >= length {
			uperIndex = length - 1
		}

		d.chunkRanges = append(d.chunkRanges, chunkRange{
			chunkId: chunk_id,
			header:  fmt.Sprintf("bytes=%d-%d", lowerIndex, uperIndex),
		})
		fmt.Println(fmt.Sprintf("Chunk %d: [%d-%d]", chunk_id, lowerIndex, uperIndex))
		chunk_id++
		lowerIndex = lowerIndex + d.options.ChunkSizeInBytes
	}

	d.firstChunkID = 1
	d.lastChunkID = chunk_id - 1

	return nil
}

func main() {

	url := "https://test-videos.co.uk/vids/bigbuckbunny/mp4/h264/1080/Big_Buck_Bunny_1080_10s_5MB.mp4"
	options := Options{
		Url:              url,
		ChunkSizeInBytes: 1024 * 1024 * 1,
		Concurrency:      5,
		Retries:          3,
		outputFilePath:   "output.mp4",
	}

	d, err := NewConcurrentDownloader(options)
	if err != nil {
		println(err.Error())
	}

	err = d.StartDownload()
	if err != nil {
		println(err.Error())
	}

}
