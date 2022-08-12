package models

type DataChunk struct {
	ChunkId      int
	Data         []byte
	DownloadTime int //ms
}
