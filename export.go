package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
)

type ManifestCommit struct {
	Hash     string   `json:"hash"`
	NoteRefs []string `json:"note_refs"`
}

type ManifestBenchmark struct {
	Name    string   `json:"name"`
	Commits []string `json:"commits"`
}

type Manifest struct {
	// Metadata information
	Module       string    `json:"module"`
	LatestCommit string    `json:"latest_commit"`
	GeneratedAt  time.Time `json:"generated_at"`

	// Benchmarks are all benchmarks found in the notes.
	Benchmarks []ManifestBenchmark `json:"benchmarks"`

	// Commits are sorted by newest to oldest.
	Commits []ManifestCommit `json:"commits"`
}

type exportPayload struct {
	output string
	data   any
}

func (cmd *cmd) Export() *cli.Command {
	return &cli.Command{
		Name:  "export",
		Usage: "Export benchmark notes to a file",
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "since", UsageText: "<git ref>"},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "single", Usage: "single commit to export (bypasses range)"},
			&cli.StringFlag{Name: "output", Value: "benchmarks", Usage: "output file path"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			since := c.StringArg("since")
			single := c.Bool("single")
			output := c.String("output")

			if since == "" {
				return fmt.Errorf("missing required argument: since")
			}

			rangeSpec := since + "^..HEAD"
			if single {
				rangeSpec = since + "^.." + since
			}

			commits, err := gitRevList(ctx, rangeSpec)
			if err != nil {
				return fmt.Errorf("git rev-list: %w", err)
			}
			if len(commits) == 0 {
				cmd.logger.InfoContext(ctx, "no commits found in range")
				return nil
			}

			cmd.logger.InfoContext(ctx, "export start", "range", rangeSpec)

			refs, err := listAllNotesRefs(ctx)
			if err != nil {
				return fmt.Errorf("listing notes refs: %w", err)
			}
			if len(refs) == 0 {
				cmd.logger.InfoContext(ctx, "no notes refs found, nothing to export")
				return nil
			}

			manifest := &Manifest{
				GeneratedAt:  time.Now().UTC(),
				LatestCommit: commits[0],
			}
			writeTasks := make(chan exportPayload, 10)

			group, errCtx := errgroup.WithContext(ctx)
			group.Go(func() error {
				return exportNote(errCtx, writeTasks)
			})

			emptyCommits := 0
			benchmarks := map[string]*ManifestBenchmark{}
			for _, commit := range commits {
				var noteRefs []string
				for _, ref := range refs {
					value, err := gitNotesShow(errCtx, ref, commit)
					if err != nil {
						continue // This commit doesn't have this note ref, skip it.
					}
					var note Note
					if err := json.Unmarshal(value, &note); err != nil {
						cmd.logger.WarnContext(errCtx, "unknown note ref", "err", err, "commit", commit, "ref", ref)
						continue
					}
					noteRefs = append(noteRefs, ref)
					select {
					case <-errCtx.Done():
						return errCtx.Err()
					case writeTasks <- exportPayload{
						output: filepath.Join(output, "commits", commit, ref+".json"),
						data:   note,
					}:
					}

					if manifest.Module == "" && note.Parsed.Pkg != "" {
						manifest.Module = note.Parsed.Pkg
					}

					for _, bench := range note.Parsed.Benches {
						bm, ok := benchmarks[bench.Name]
						if !ok {
							bm = &ManifestBenchmark{Name: bench.Name, Commits: []string{commit}}
							benchmarks[bench.Name] = bm
						} else {
							if slices.Contains(bm.Commits, commit) {
								continue
							}
							bm.Commits = append(bm.Commits, commit)
						}
						select {
						case <-errCtx.Done():
							return errCtx.Err()
						case writeTasks <- exportPayload{
							output: filepath.Join(output, "benchmarks", bench.Name, commit+".json"),
							data:   bench,
						}:
						}
					}
				}
				if len(noteRefs) == 0 {
					emptyCommits++
					continue // No notes for this commit, skip it.
				}
				manifest.Commits = append(manifest.Commits, ManifestCommit{
					Hash:     commit,
					NoteRefs: noteRefs,
				})
			}

			close(writeTasks)
			if err := group.Wait(); err != nil {
				return fmt.Errorf("exporting notes: %w", err)
			}

			for _, bm := range benchmarks {
				manifest.Benchmarks = append(manifest.Benchmarks, *bm)
			}

			if err := writeManifest(manifest, output); err != nil {
				return fmt.Errorf("writing manifest: %w", err)
			}

			cmd.logger.InfoContext(ctx, "export complete", "commits", len(manifest.Commits), "empty_commits", emptyCommits, "notes", len(manifest.Commits), "benchmarks", len(manifest.Benchmarks))

			return nil
		},
	}
}

func exportNote(ctx context.Context, tasks <-chan exportPayload) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case task, ok := <-tasks:
			if !ok {
				return ctx.Err()
			}
			if err := os.MkdirAll(filepath.Dir(task.output), 0o755); err != nil {
				return fmt.Errorf("creating directory %s: %w", filepath.Dir(task.output), err)
			}
			data, err := json.Marshal(task.data)
			if err != nil {
				return fmt.Errorf("marshaling export data for %s: %w", task.output, err)
			}
			if err := os.WriteFile(task.output, data, 0o644); err != nil {
				return fmt.Errorf("writing export file at %s: %w", task.output, err)
			}
		}
	}
}

func writeManifest(manifest *Manifest, output string) error {
	manifestPath := filepath.Join(output, "manifest.json")
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		return fmt.Errorf("mkdir at %s: %w", filepath.Dir(manifestPath), err)
	}
	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling manifest: %w", err)
	}
	if err := os.WriteFile(manifestPath, manifestJSON, 0o644); err != nil {
		return fmt.Errorf("writing manifest at %s: %w", manifestPath, err)
	}
	return nil
}
