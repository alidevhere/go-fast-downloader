package downloader

import (
	"errors"
	"time"
)

var (
	HeaderRangesNotSupported = errors.New("Header ranges not supported")
)

const (
	TimeOutDuration = 10 * time.Second
)
