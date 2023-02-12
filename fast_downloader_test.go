package fastdownloader

import (
	"net/http"
	"sync"
	"testing"
	"time"
)

func Test_validateOpts(t *testing.T) {
	type args struct {
		options *Options
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid url",
			args: args{
				options: &Options{
					Url: "",
				}},
			wantErr: true,
		},
		{
			name: "invalid OutputFileName",
			args: args{
				options: &Options{
					Url: "http://www.google.com",
				}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateOpts(tt.args.options); (err != nil) != tt.wantErr {
				t.Errorf("validateOpts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_downloader_OutputFilePath(t *testing.T) {
	type fields struct {
		firstChunkID   int
		lastChunkID    int
		client         *http.Client
		options        Options
		chunkRanges    []chunkRange
		wg             *sync.WaitGroup
		inputChan      chan chunkRange
		outputChan     chan dataChunk
		stopChan       chan struct{}
		outputFilePath string
		downloadTime   time.Duration
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "output file path test-1",
			fields: fields{
				outputFilePath: "",
			},
			want: "",
		},
		{
			name: "output file path test-2",
			fields: fields{
				outputFilePath: "abc/a.txt",
			},
			want: "abc/a.txt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &downloader{
				firstChunkID:   tt.fields.firstChunkID,
				lastChunkID:    tt.fields.lastChunkID,
				client:         tt.fields.client,
				options:        tt.fields.options,
				chunkRanges:    tt.fields.chunkRanges,
				wg:             tt.fields.wg,
				inputChan:      tt.fields.inputChan,
				outputChan:     tt.fields.outputChan,
				stopChan:       tt.fields.stopChan,
				outputFilePath: tt.fields.outputFilePath,
				downloadTime:   tt.fields.downloadTime,
			}
			if got := d.OutputFilePath(); got != tt.want {
				t.Errorf("downloader.OutputFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_downloader_DownloadTime(t *testing.T) {
	type fields struct {
		firstChunkID   int
		lastChunkID    int
		client         *http.Client
		options        Options
		chunkRanges    []chunkRange
		wg             *sync.WaitGroup
		inputChan      chan chunkRange
		outputChan     chan dataChunk
		stopChan       chan struct{}
		outputFilePath string
		downloadTime   time.Duration
	}
	tests := []struct {
		name   string
		fields fields
		want   time.Duration
	}{
		{
			name: "download time test-1",
			fields: fields{
				downloadTime: 0,
			},
			want: 0,
		},
		{
			name: "download time test-2",
			fields: fields{
				downloadTime: 100 * time.Second,
			},
			want: 100 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &downloader{
				firstChunkID:   tt.fields.firstChunkID,
				lastChunkID:    tt.fields.lastChunkID,
				client:         tt.fields.client,
				options:        tt.fields.options,
				chunkRanges:    tt.fields.chunkRanges,
				wg:             tt.fields.wg,
				inputChan:      tt.fields.inputChan,
				outputChan:     tt.fields.outputChan,
				stopChan:       tt.fields.stopChan,
				outputFilePath: tt.fields.outputFilePath,
				downloadTime:   tt.fields.downloadTime,
			}
			if got := d.DownloadTime(); got != tt.want {
				t.Errorf("downloader.DownloadTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_downloader_calculateChunkRanges(t *testing.T) {
	type fields struct {
		firstChunkID   int
		lastChunkID    int
		client         *http.Client
		options        Options
		chunkRanges    []chunkRange
		wg             *sync.WaitGroup
		inputChan      chan chunkRange
		outputChan     chan dataChunk
		stopChan       chan struct{}
		outputFilePath string
		downloadTime   time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &downloader{
				firstChunkID:   tt.fields.firstChunkID,
				lastChunkID:    tt.fields.lastChunkID,
				client:         tt.fields.client,
				options:        tt.fields.options,
				chunkRanges:    tt.fields.chunkRanges,
				wg:             tt.fields.wg,
				inputChan:      tt.fields.inputChan,
				outputChan:     tt.fields.outputChan,
				stopChan:       tt.fields.stopChan,
				outputFilePath: tt.fields.outputFilePath,
				downloadTime:   tt.fields.downloadTime,
			}
			if err := d.calculateChunkRanges(); (err != nil) != tt.wantErr {
				t.Errorf("downloader.calculateChunkRanges() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
