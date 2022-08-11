package models

type ChunkInfo struct {
	ChunkId    int
	Uri        string
	ChunkRange string
}

type ChunkRange struct {
	LowerRange int
	UpperRange int
}
