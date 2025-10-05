package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"golang.org/x/perf/benchfmt"
)

// Note represents the benchmark note stored as a git note.
type Note struct {
	Schema    int          `json:"schema"`
	Commit    string       `json:"commit"`
	CreatedAt time.Time    `json:"created_at"`
	Env       EnvInfo      `json:"env"`
	BenchArgs []string     `json:"bench_args"`
	Parsed    *BenchReport `json:"parsed"`
	Raw       string       `json:"raw"`
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
	parsed, err := parseBenchReport(raw)
	if err != nil {
		return nil, fmt.Errorf("parsing bench report: %w", err)
	}
	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
	}
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
		Parsed:    parsed,
	})
}

// parseBenchReport parses raw `go test -bench` output into a BenchReport.
// It uses benchfmt for benchmark lines and a light pre-pass for headers.
func parseBenchReport(raw []byte) (*BenchReport, error) {
	rep := &BenchReport{}
	parseHeadersAndFooter(rep, raw)

	// map from benchmark name to *BenchCase
	byName := map[string]*BenchCase{}

	rdr := benchfmt.NewReader(bytes.NewReader(raw), "bench")
	for rdr.Scan() {
		rec := rdr.Result()
		switch rec := rec.(type) {
		case *benchfmt.Result:
			name := rec.Name.String()
			c := byName[name]
			if c == nil {
				c = &BenchCase{Name: name}
				byName[name] = c
			}
			s := BenchSample{
				Iterations: int64(rec.Iters),
			}
			// time/op (normalize to ns)
			if v, ok := valueNsPerOp(rec); ok {
				s.NsPerOp = v
			}
			// bytes/op, allocs/op
			if v, ok := rec.Value("B/op"); ok {
				s.BytesPerOp = int64(v)
			}
			if v, ok := rec.Value("allocs/op"); ok {
				s.AllocsPerOp = int64(v)
			}
			c.Samples = append(c.Samples, s)
		default:
			// ignore other record types (syntax errors, config, etc.)
		}
	}
	if err := rdr.Err(); err != nil {
		return nil, fmt.Errorf("benchfmt reader error: %w", err)
	}

	// convert map to slice and compute stats
	rep.Benches = make([]BenchCase, 0, len(byName))
	for _, bc := range byName {
		sort.Slice(bc.Samples, func(i, j int) bool {
			a, b := bc.Samples[i], bc.Samples[j]
			if a.Iterations != b.Iterations {
				return a.Iterations > b.Iterations
			}
			return a.NsPerOp < b.NsPerOp
		})
		bc.Stats = computeBenchStats(bc.Samples)
		rep.Benches = append(rep.Benches, *bc)
	}
	sort.Slice(rep.Benches, func(i, j int) bool {
		return rep.Benches[i].Name < rep.Benches[j].Name
	})

	return rep, nil
}

// valueNsPerOp returns time/op normalized to nanoseconds, regardless of the
// unit used in the benchmark output (ns/op, µs/op, ms/op, s/op).
func valueNsPerOp(rec *benchfmt.Result) (float64, bool) {
	// Try common time units in descending precision.
	if v, ok := rec.Value("ns/op"); ok {
		return v, true
	}
	// Some terminals/files use the micro sign U+00B5; others may have ASCII "us/op".
	if v, ok := rec.Value("µs/op"); ok { // U+00B5
		return v * 1e3, true // µs -> ns
	}
	if v, ok := rec.Value("us/op"); ok { // ASCII fallback
		return v * 1e3, true
	}
	if v, ok := rec.Value("ms/op"); ok {
		return v * 1e6, true // ms -> ns
	}
	if v, ok := rec.Value("s/op"); ok {
		return v * 1e9, true // s -> ns
	}
	if v, ok := rec.Value("sec/op"); ok {
		return v * 1e9, true // s -> ns
	}
	return 0, false
}

var (
	reKV     = regexp.MustCompile(`^\s*([a-z]+):\s*(.+?)\s*$`)       // goos:, goarch:, pkg:, cpu:
	reOKLine = regexp.MustCompile(`^\s*ok\s+(\S+)\s+([0-9.]+s)\s*$`) // ok <pkg> 12.916s
	rePkg    = regexp.MustCompile(`^\s*pkg:\s*(\S+)\s*$`)            // pkg: github.com/...
)

func parseHeadersAndFooter(rep *BenchReport, raw []byte) {
	sc := bufio.NewScanner(bytes.NewReader(raw))
	for sc.Scan() {
		line := sc.Text()

		// ok <pkg> <dur>
		if m := reOKLine.FindStringSubmatch(line); m != nil {
			rep.Pkg = m[1]
			if d, err := time.ParseDuration(m[2]); err == nil {
				rep.Duration = d
			}
			rep.Pass = true
			continue
		}

		// "pkg: <path>" (some Go versions emit this form too)
		if m := rePkg.FindStringSubmatch(line); m != nil {
			rep.Pkg = m[1]
			continue
		}

		// Generic k:v header lines (goos, goarch, cpu, etc.)
		if m := reKV.FindStringSubmatch(line); m != nil {
			k, v := m[1], strings.TrimSpace(m[2])
			switch k {
			case "goos":
				rep.GOOS = v
			case "goarch":
				rep.GOARCH = v
			case "pkg":
				rep.Pkg = v
			case "cpu":
				rep.CPU = squeezeSpaces(v)
			}
		}
	}
	_ = sc.Err() // ignore; headers are best-effort
}

func squeezeSpaces(s string) string {
	fs := strings.Fields(s)
	return strings.Join(fs, " ")
}

// BenchReport corresponds to a single `go test -bench` output file.
type BenchReport struct {
	GOOS     string        `json:"goos,omitempty"`   // e.g., "linux"
	GOARCH   string        `json:"goarch,omitempty"` // e.g., "amd64"
	Pkg      string        `json:"pkg,omitempty"`    // e.g., "github.com/zackarysantana/gb"
	CPU      string        `json:"cpu,omitempty"`    // optional: parsed from header if present
	Pass     bool          `json:"pass"`             // true if "PASS" seen
	Duration time.Duration `json:"duration"`         // e.g., "12.916s" at end (ok … 12.916s)

	// One entry per benchmark function (e.g., BenchmarkConcatJoin)
	Benches []BenchCase `json:"benches"`
}

// BenchCase aggregates all repeated lines for one benchmark name.
type BenchCase struct {
	Name    string        `json:"name"`    // e.g., "BenchmarkConcatJoin-24"
	Samples []BenchSample `json:"samples"` // one per printed line
	Stats   BenchStats    `json:"stats"`   // computed summary across samples
}

// BenchSample is a single printed benchmark line (one repetition).
// Use float64 for rates/times to make stats easy; keep ints for counters.
type BenchSample struct {
	Iterations  int64   `json:"iterations"`              // e.g., 123530
	NsPerOp     float64 `json:"ns_per_op"`               // e.g., 9670
	BytesPerOp  int64   `json:"bytes_per_op,omitempty"`  // e.g., 22528
	AllocsPerOp int64   `json:"allocs_per_op,omitempty"` // e.g., 2
}

// BenchStats is a compact summary over all Samples for a BenchCase.
type BenchStats struct {
	Count int `json:"count"`
	// Time
	NsPerOpMean   float64 `json:"ns_per_op_mean"`
	NsPerOpMedian float64 `json:"ns_per_op_median"`
	NsPerOpMin    float64 `json:"ns_per_op_min"`
	NsPerOpMax    float64 `json:"ns_per_op_max"`
	// Memory & allocs (use -benchmem to populate)
	BytesPerOpMean    float64 `json:"bytes_per_op_mean,omitempty"`
	BytesPerOpMedian  float64 `json:"bytes_per_op_median,omitempty"`
	AllocsPerOpMean   float64 `json:"allocs_per_op_mean,omitempty"`
	AllocsPerOpMedian float64 `json:"allocs_per_op_median,omitempty"`
}

func computeBenchStats(samples []BenchSample) BenchStats {
	var s BenchStats
	n := len(samples)
	s.Count = n
	if n == 0 {
		return s
	}

	ns := make([]float64, 0, n)
	var nsMin, nsMax, nsSum float64
	nsMin = 1<<63 - 1
	nsMax = -1

	var bytesVals []float64
	var allocsVals []float64
	var bytesSum, allocsSum float64

	for _, sm := range samples {
		ns = append(ns, sm.NsPerOp)
		nsSum += sm.NsPerOp
		if sm.NsPerOp < nsMin {
			nsMin = sm.NsPerOp
		}
		if sm.NsPerOp > nsMax {
			nsMax = sm.NsPerOp
		}
		if sm.BytesPerOp > 0 {
			v := float64(sm.BytesPerOp)
			bytesVals = append(bytesVals, v)
			bytesSum += v
		}
		if sm.AllocsPerOp > 0 {
			v := float64(sm.AllocsPerOp)
			allocsVals = append(allocsVals, v)
			allocsSum += v
		}
	}

	sort.Float64s(ns)
	s.NsPerOpMean = nsSum / float64(len(ns))
	s.NsPerOpMedian = median(ns)
	s.NsPerOpMin = nsMin
	s.NsPerOpMax = nsMax

	if len(bytesVals) > 0 {
		sort.Float64s(bytesVals)
		s.BytesPerOpMean = bytesSum / float64(len(bytesVals))
		s.BytesPerOpMedian = median(bytesVals)
	}
	if len(allocsVals) > 0 {
		sort.Float64s(allocsVals)
		s.AllocsPerOpMean = allocsSum / float64(len(allocsVals))
		s.AllocsPerOpMedian = median(allocsVals)
	}
	return s
}

func median(sorted []float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2
}
