package downloader

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	HeaderRangesNotSupported = errors.New("Header ranges not supported")
)

const (
	TimeOutDuration = 20 * time.Second
)

type DownloadOptions struct {
	//Download file in this many chunks
	NoOfChunks int
	//No of go routines in pool
	NoOfWorkers int
	//Url
	URI string
}

//Start Pool make live all routines

func Download(c DownloadOptions) {

	chunkRange, err := calculateChunkSlice(c.URI, c.NoOfChunks)
	logInfo("[MANAGER]: Calculated Chunk Range : " + fmt.Sprintf("%v", chunkRange))
	if err == HeaderRangesNotSupported {
		log.Printf("Error: %v", err.Error())
		panic(err)

	}

	if err != nil {
		log.Printf("Error: %v", err.Error())
		panic(err)
	}

	//Chunk Ranges is supported so:
	//1- Make workers pool
	//2- assign chunks

	chunkInfoChan := make(chan ChunkInfo, 100)
	stopChan := make(chan struct{})
	Output := make(chan DataChunk, 100)
	wg := new(sync.WaitGroup)

	for i := 1; i <= c.NoOfWorkers; i++ {
		w := Worker{
			WorkerId:  i,
			ChunkInfo: chunkInfoChan,
			Stop:      stopChan,
			Wg:        wg,
			Output:    Output,
		}
		wg.Add(1)
		go w.StartDownload()
	}

	//Create Merger
	merger := Merger{
		TotalWorkers:      c.NoOfWorkers,
		FirstId:           1,
		LastId:            len(chunkRange),
		Input:             Output,
		File:              "test.tmp",
		Wg:                wg,
		StopReceivingChan: stopChan,
	}
	go merger.Start()
	wg.Add(1)

	for _, i := range chunkRange {
		chunkInfoChan <- ChunkInfo{
			ChunkId:    i.ChunkId,
			Uri:        c.URI,
			ChunkRange: fmt.Sprintf("bytes=%d-%d", i.LowerRange, i.UpperRange),
		}
	}
	logSuccess("[MANAGER]: Sent total chunksInfo to download %d first Chunk ID[%d] last chunk ID[%d]", len(chunkRange), chunkRange[0].ChunkId, chunkRange[len(chunkRange)-1].ChunkId)

	wg.Wait()

}

func calculateChunkSlice(uri string, chunkSize int) ([]ChunkRange, error) {
	var cr []ChunkRange
	r, err := http.NewRequest("HEAD", uri, nil)
	if err != nil {
		return cr, err
	}

	client := http.Client{
		Timeout: TimeOutDuration,
	}

	res, err := client.Do(r)
	if err != nil {
		return cr, err
	}

	if res.StatusCode != 200 && res.StatusCode != 206 {
		log.Printf("Status Code (%d): Invalid header received\n", res.StatusCode)
		return cr, HeaderRangesNotSupported
	}
	defer res.Body.Close()

	cl := res.Header.Get("content-length")
	logSuccess("[MANAGER]: Calculated Content Length : " + cl)
	length, err := strconv.Atoi(cl)
	if err != nil {
		return cr, err
	}

	//it will be 'none' if ranges not supported
	if res.Header.Get("Accept-Ranges") != "bytes" {
		return cr, HeaderRangesNotSupported
	}

	//Ranges supported
	chunk_id := 1
	chunkLength := length / chunkSize
	for i := 0; i < length; i = i + chunkLength + 1 {
		ul := i + chunkLength
		if ul > length {
			ul = length
		}

		cr = append(cr, ChunkRange{
			ChunkId:    chunk_id,
			LowerRange: i,
			UpperRange: ul,
		})
		chunk_id++
	}
	return cr, nil
}
