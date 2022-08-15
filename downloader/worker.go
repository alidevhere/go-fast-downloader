package downloader

import (
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/alidevhere/go-fast-downloader/util"
)

// a single goroutine in pool
type Worker struct {
	WorkerId  int
	ChunkInfo chan ChunkInfo
	Output    chan DataChunk
	Stop      chan struct{}
	Err       chan struct{}
	Wg        *sync.WaitGroup
	client    http.Client
	// Status
}

//Downloads file chunk and save in global registry

// Start Download
func (w Worker) StartDownload() {
	defer w.Wg.Done()
	//Use same client for one go routine during its life cycle
	w.client = http.Client{Timeout: util.TimeOutDuration}
	for {
		select {
		case info := <-w.ChunkInfo:

			data, err := w.DownloadChunk(info)
			logSuccess("WORKER[%d]: Downloaded Chunk id[%d] Chunk Range[%s]\n", w.WorkerId, info.ChunkId, info.ChunkRange)
			if err != nil {
				log.Printf("ID[%d]:Chunk [%s] ----> FAILED \n", w.WorkerId, info.ChunkRange)
				log.Printf("ID[%d]:Retrying Chunk [%s] \n", w.WorkerId, info.ChunkRange)
				//TODO:
				//Divide into smaller chunk if failed
				w.ChunkInfo <- info
			}

			logInfo("WORKER[%d]: sending chunk %d to output channel \n", w.WorkerId, data.ChunkId)
			w.Output <- data
		case <-w.Stop:
			logSuccess("ID[%d]: Process completed logging out.....\n", w.WorkerId)
			return
			// default:
			// 	log.Printf("ID[%d]Waiting for job\nSleeping for 5 sec\n", w.WorkerId)
			// 	time.Sleep(5 * time.Second)
			// default:
			// 	log.Printf("Worker [%d]: Waiting \n", w.WorkerId)
		}
	}
}

// Start Download
func (w Worker) DownloadChunk(info ChunkInfo) (DataChunk, error) {
	r, err := http.NewRequest("GET", info.Uri, nil)
	var data DataChunk
	if err != nil {
		return data, err
	}

	r.Header.Add("Range", info.ChunkRange)
	res, err := w.client.Do(r)
	if err != nil {
		return data, err
	}
	defer res.Body.Close()

	data.Data, err = ioutil.ReadAll(res.Body)
	data.ChunkId = info.ChunkId
	if err != nil {
		return data, err
	}

	return data, nil
}
