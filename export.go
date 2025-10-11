package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
)

type ManifestCommit struct {
	Hash     string   `json:"hash"`
	NoteRefs []string `json:"note_refs"`
}

type Manifest struct {
	// Metadata information
	GeneratedAt  time.Time `json:"generated_at"`
	LatestCommit string    `json:"latest_commit"`
	Module       string    `json:"module"`

	// Commits are sorted by newest to oldest.
	Commits []ManifestCommit `json:"commits"`
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

			cmd.logger.InfoContext(ctx, "sync start", "range", rangeSpec)

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
				Module:       "TODO-MODULE-NAME",
			}

			type noteInfo struct {
				fileName string
				value    []byte
			}
			ch := make(chan noteInfo, 10)
			group, errCtx := errgroup.WithContext(ctx)

			group.Go(func() error {
				for {
					select {
					case <-errCtx.Done():
						return errCtx.Err()
					case n, ok := <-ch:
						if !ok {
							return errCtx.Err()
						}
						if err := os.MkdirAll(filepath.Dir(n.fileName), 0o755); err != nil {
							return fmt.Errorf("mkdir at %s: %w", filepath.Dir(n.fileName), err)
						}
						if err := os.WriteFile(n.fileName, n.value, 0o644); err != nil {
							return fmt.Errorf("writing note at %s: %w", n.fileName, err)
						}
						cmd.logger.DebugContext(errCtx, "wrote note", "fileName", n.fileName)
					}
				}
			})

			notesWrote := 0
			for _, commit := range commits {
				var notes []string
				for _, ref := range refs {
					value, err := gitNotesShow(ctx, ref, commit)
					if err != nil {
						// This commit doesn't have this note ref, so we skip it.
						continue
					}
					select {
					case <-ctx.Done():
						close(ch)
						group.Wait()
						return ctx.Err()
					case ch <- noteInfo{
						fileName: filepath.Join(output, commit, ref+".json"),

						value: value,
					}:
						notes = append(notes, ref)
						notesWrote++
					}
				}
				if len(notes) > 0 {
					manifest.Commits = append(manifest.Commits, ManifestCommit{
						Hash:     commit,
						NoteRefs: notes,
					})
				}
			}

			close(ch)
			if err := group.Wait(); err != nil {
				return fmt.Errorf("exporting notes: %w", err)
			}

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

			commitsWithoutNotes := len(commits) - len(manifest.Commits)
			cmd.logger.InfoContext(ctx, "export complete", "range", rangeSpec, "commits", len(commits), "notes", notesWrote, "commitsWithoutNotes", commitsWithoutNotes)
			return nil
		},
	}
}
