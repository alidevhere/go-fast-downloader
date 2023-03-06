package fastdownloader

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// Options for creating new the new instance downloader.
type Options struct {
	// The number of bytes to download in each request.
	// If not set, the default is 5MB or 1024*1024*5 Bytes.
	ChunkSizeInBytes int
	// The number of concurrent requests to make.
	// If not set, the default is 5.
	Concurrency int
	// The number of retries to make for each request.
	// If not set, the default is 3.
	Retries int
	// Url to download from.
	Url string
	// Request timeout in seconds.
	// If not set, the default is 3 minutes.
	RequestTimeout time.Duration
	// Output file name.
	OutputFileName string
	// Output file directory.
	// default output directory is the current user's home/downloads directory.
	OutputFileDirectory string
}

type ConcurrentDownloader interface {
	// Start the download process.
	StartDownload() error
	// Get the output file path.
	OutputFilePath() string
	// Get the total download time.
	DownloadTime() time.Duration

	// IsDownloadComplete() bool
}

type downloader struct {
	firstChunkID   int
	lastChunkID    int
	client         *http.Client
	options        Options
	chunkRanges    []chunkRange
	wg             *sync.WaitGroup
	inputChan      chan chunkRange
	outputChan     chan dataChunk
	stopChan       chan struct{}
	outputFilePath string
	downloadTime   time.Duration
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

	f, err := os.CreateTemp(d.options.OutputFileDirectory, "temp.tmp")
	if err != nil {
		log.Println(err)
	}
	defer os.Remove(f.Name())

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
				of, e := os.OpenFile(filepath.Join(d.options.OutputFileDirectory, d.options.OutputFileName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
				if e != nil {
					println(e.Error())
				}

				defer of.Close()
				f.Close()
				sortOutputFiles(m, d.firstChunkID, d.lastChunkID, f.Name(), of)
				d.outputFilePath = of.Name()
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
	if err := validateOpts(&options); err != nil {
		return nil, err
	}

	return &downloader{options: options, client: &http.Client{
		Timeout: options.RequestTimeout,
	}}, nil

}

func validateOpts(options *Options) error {
	if options.ChunkSizeInBytes == 0 {
		options.ChunkSizeInBytes = 1024 * 1024 * 1
	}

	if options.Concurrency == 0 {
		options.Concurrency = 5
	}

	if options.Retries == 0 {
		options.Retries = 3
	}

	if options.Url == "" {
		return errors.New("Url is required")
	}

	if options.OutputFileDirectory == "" {
		return errors.New("Output file directory is required")
		// usr, err := user.Current()
		// if err != nil {
		// 	return err
		// }
		// options.OutputFileDirectory = filepath.Join(usr.HomeDir, "Downloads")
	}

	if options.OutputFileName == "" {
		return errors.New("Output file name is required")
	}

	if options.RequestTimeout == 0 {
		println("Setting request timeout to 3 minutes")
		options.RequestTimeout = 3 * time.Minute
	}

	return nil
}

func (d *downloader) OutputFilePath() string {
	return d.outputFilePath
}

func (d *downloader) DownloadTime() time.Duration {
	return d.downloadTime
}

func (d *downloader) StartDownload() error {

	startTime := time.Now()
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
	close(d.inputChan)
	close(d.outputChan)
	close(d.stopChan)

	d.downloadTime = time.Since(startTime)

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
