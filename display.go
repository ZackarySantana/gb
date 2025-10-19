package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"

	"github.com/urfave/cli/v3"
)

//go:embed public/*
var webFS embed.FS

func (cmd *cmd) Display() *cli.Command {
	return &cli.Command{
		Name:  "display",
		Usage: "Display benchmark notes via a UI",
		Action: func(ctx context.Context, c *cli.Command) error {
			benchmarkDirectory := "./benchmarks"

			staticFS, err := fs.Sub(webFS, "public")
			if err != nil {
				return fmt.Errorf("getting sub filesystem: %w", err)
			}

			mux := http.NewServeMux()

			mux.Handle("/data/", http.StripPrefix("/data/",
				http.FileServer(http.Dir(benchmarkDirectory)),
			))

			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				reqPath := r.URL.Path
				if reqPath == "/" {
					reqPath = "index.html"
				} else {
					reqPath = reqPath[1:]
				}

				f, err := staticFS.Open(reqPath)
				if err != nil {
					// fallback to index.html (for SPAs)
					f, err = staticFS.Open("index.html")
					if err != nil {
						http.NotFound(w, r)
						return
					}
				}
				defer f.Close()

				info, _ := f.Stat()

				rs, ok := f.(io.ReadSeeker)
				if !ok {
					data, err := io.ReadAll(f)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					http.ServeContent(w, r, path.Base(reqPath), info.ModTime(), bytesReader(data))
					return
				}

				http.ServeContent(w, r, path.Base(reqPath), info.ModTime(), rs)
			})

			fmt.Println("Serving embedded site on", "http://localhost:8080")
			return http.ListenAndServe(":8080", mux)
		},
	}
}

func bytesReader(b []byte) io.ReadSeeker {
	return &reader{b: b}
}

type reader struct {
	b []byte
	i int64
}

func (r *reader) Read(p []byte) (int, error) {
	if r.i >= int64(len(r.b)) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.i:])
	r.i += int64(n)
	return n, nil
}

func (r *reader) Seek(offset int64, whence int) (int64, error) {
	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = r.i + offset
	case io.SeekEnd:
		newPos = int64(len(r.b)) + offset
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}
	if newPos < 0 {
		return 0, fmt.Errorf("invalid seek position")
	}
	r.i = newPos
	return newPos, nil
}
