package downloader

type ChunkInfo struct {
	ChunkId    int
	Uri        string
	ChunkRange string
}

type ChunkRange struct {
	LowerRange int
	UpperRange int
}

type DataChunk struct {
	ChunkId      int
	Data         []byte
	DownloadTime int //ms
}
