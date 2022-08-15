package downloader

type DataChunk struct {
	ChunkId      int
	Data         []byte
	DownloadTime int //ms
}
