package models

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/alidevhere/go-fast-downloader/util"
)

type DownloadOptions struct {
	//chunk size in bytes
	ChunkSize int
	//No of go routines in pool
	NoOfWorkers int
	//Url
	URI string
}

//Start Pool make live all routines

func Download(c DownloadOptions) {

	chunkRange, err := calculateChunkSlice(c.URI, c.ChunkSize)
	logInfo("[MANAGER]: Calculated Chunk Range : " + fmt.Sprintf("%v", chunkRange))
	if err == HeaderRangesNotSupported {
		//TODO:
		//Download file with single file and without pause functionality
		log.Printf("Error: %v", err.Error())
		panic(err)

	}

	if err != nil {
		log.Printf("Error: %v", err.Error())
		panic(err)
	}

	//Chunk Ranges is supported so:
	//1- Make pool
	//2- assign chunks

	chunkInfoChan := make(chan ChunkInfo)
	stopChan := make(chan struct{})
	Output := make(chan DataChunk)
	wg := new(sync.WaitGroup)
	//Saving data in Pool
	// wp := WorkerPool{
	// 	NoOfWorkers:    c.NoOfWorkers,
	// 	InputChan:      chunkInfoChan,
	// 	StopChan:       stopChan,
	// 	Wg:             wg,
	// 	ChunksRegistry: make(map[string][]byte),
	// }

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

	for id, i := range chunkRange {
		chunkInfoChan <- ChunkInfo{
			ChunkId:    id + 1,
			Uri:        c.URI,
			ChunkRange: fmt.Sprintf("bytes=%d-%d", i.LowerRange, i.UpperRange),
		}
	}
	logSuccess("[MANAGER]: Sent total chunksInfo to download %d", len(chunkRange))

	wg.Wait()

}

func calculateChunkSlice(uri string, chunkSize int) ([]ChunkRange, error) {
	var cr []ChunkRange
	r, err := http.NewRequest("HEAD", uri, nil)
	if err != nil {
		return cr, err
	}

	client := http.Client{
		Timeout: util.TimeOutDuration,
	}
	res, err := client.Do(r)
	//ioutil.ReadAll(res.Body)
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

	//it will be 'none' is ranges not supported
	if res.Header.Get("Accept-Ranges") != "bytes" {
		return cr, HeaderRangesNotSupported
	}

	//Ranges supported

	chunkLength := length / chunkSize
	for i := 0; i < length; i = i + chunkLength + 1 {
		ul := i + chunkLength
		if ul > length {
			ul = length
		}

		cr = append(cr, ChunkRange{
			LowerRange: i,
			UpperRange: ul,
		})

	}
	return cr, nil
}

// This function will check if on channel if passed chunk is nect required chunk then
// Save in file otherwise save in map chunkRegistry
// func chunksMerger(c WorkerPoolOptions) (WorkerPool, error) {

// }
