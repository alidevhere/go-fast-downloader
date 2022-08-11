package models

import (
	"log"
	"net/http"
	"sync"

	"github.com/alidevhere/go-fast-downloader/util"
)

// a single goroutine in pool
type Worker struct {
	WorkerId  int
	ChunkInfo chan ChunkInfo
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
			log.Printf("ID[%d]: Chunk [%s]\n", w.WorkerId, info.ChunkRange)

			if err := w.DownloadChunk(info); err != nil {
				log.Printf("ID[%d]:Chunk [%s] ----> FAILED \n", w.WorkerId, info.ChunkRange)
				log.Printf("ID[%d]:Retrying Chunk [%s] \n", w.WorkerId, info.ChunkRange)
				//TODO:
				//Divide into smaller chunk if failed
				w.ChunkInfo <- info
			}
		case <-w.Stop:
			log.Printf("ID[%d]: Process completed logging out.....\n", w.WorkerId)
			// default:
			// 	log.Printf("ID[%d]Waiting for job\nSleeping for 5 sec\n", w.WorkerId)
			// 	time.Sleep(5 * time.Second)
		}
	}
}

// Start Download
func (w Worker) DownloadChunk(info ChunkInfo) error {
	r, err := http.NewRequest("GET", info.Uri, nil)
	//
	if err != nil {
		return err
	}

	r.Header.Add("Range", info.ChunkRange)
	res, err := w.client.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	// data, err := ioutil.ReadAll(res.Body)
	// if err != nil {
	// 	return err
	// }
	return nil
}
