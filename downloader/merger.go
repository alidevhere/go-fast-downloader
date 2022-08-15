package downloader

import (
	"log"
	"os"
	"sync"
)

type Merger struct {
	TotalWorkers      int
	Wg                *sync.WaitGroup
	FirstId           int
	LastId            int
	currentId         int
	chunksLeft        int
	StopReceivingChan chan struct{}
	Input             chan DataChunk
	File              string
	chunkRepo         map[int][]byte
}

func (m *Merger) Start() {
	defer m.Wg.Done()
	logInfo("[MERGER]: Starting merger")
	//For saving data
	m.chunkRepo = make(map[int][]byte)
	m.currentId = m.FirstId
	//m.receivedChunks = m.LastId - m.FirstId + 1
	m.chunksLeft = m.LastId
	f, err := os.OpenFile(m.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	for {

		select {
		//Chunk of ChunkId can be in one of 3 stages
		//1. Received from channel
		//2. Present in chunkRepo
		//3. Neither of above, being downloaded

		case chunk := <-m.Input:
			logInfo("[MERGER]: Received chunk %d", chunk.ChunkId)
			m.chunksLeft--
			//merge into file
			if chunk.ChunkId == m.currentId {
				f.Write(chunk.Data)
				m.currentId++
			} else {
				m.chunkRepo[chunk.ChunkId] = chunk.Data
			}

			if data, ok := m.chunkRepo[m.currentId]; ok {
				logInfo("[MERGER]: Merged chunk %d", m.currentId)
				f.Write(data)
				//delete data from chunkRepo
				delete(m.chunkRepo, m.currentId)
				m.currentId++
			}
			//stop when last chunk is received and merge all chunks
		default:
			if m.chunksLeft == 0 {
				logInfo("[MERGER]: Received all chunks")
				//Send signle to all worker to stop
				for i := 0; i < m.TotalWorkers; i++ {

					m.StopReceivingChan <- struct{}{}
				}

				for m.currentId <= m.LastId {
					if data, ok := m.chunkRepo[m.currentId]; ok {
						f.Write(data)
						logSuccess("[MERGER]: Merged chunk %d", m.currentId)
						delete(m.chunkRepo, m.currentId)
						m.currentId++
					}
				}
				logSuccess("[MERGER]: Download Completed.... ")
				return
			}
			//log.Printf("[MERGER]: Waiting for left chunks %d", m.chunksLeft)

		}
	}

}

// Append given chunk to file
func (m *Merger) sortAndMerge() {

}
