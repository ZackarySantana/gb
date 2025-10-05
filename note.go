package main

import (
	"encoding/json"
	"os"
	"runtime"
	"time"
)

// Note represents the benchmark note stored as a git note.
type Note struct {
	Schema    int       `json:"schema"`
	Commit    string    `json:"commit"`
	CreatedAt time.Time `json:"created_at"`
	Env       EnvInfo   `json:"env"`
	BenchArgs []string  `json:"bench_args"`
	Raw       string    `json:"raw"`
}

// EnvInfo describes the Go runtime and host environment at benchmark time.
type EnvInfo struct {
	GoVersion  string `json:"go_version"`
	GoOS       string `json:"goos"`
	GoArch     string `json:"goarch"`
	GoMaxProcs int    `json:"gomaxprocs"`
	Host       string `json:"host"`
	CPUs       int    `json:"cpus"`
}

func marshalNotePayload(commit string, benchArgs []string, raw []byte) ([]byte, error) {
	host, _ := os.Hostname()
	return json.Marshal(Note{
		Schema:    1,
		Commit:    commit,
		CreatedAt: time.Now().UTC(),
		Env: EnvInfo{
			GoVersion:  runtime.Version(),
			GoOS:       runtime.GOOS,
			GoArch:     runtime.GOARCH,
			GoMaxProcs: runtime.GOMAXPROCS(0),
			Host:       host,
			CPUs:       runtime.NumCPU(),
		},
		BenchArgs: benchArgs,
		Raw:       string(raw),
	})
}
