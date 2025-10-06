# Go Benchmark Notes Manager

`gb` stores benchmarks in git notes, making it easy to track performance changes over time.

## Installation

```bash
go install github.com/zackarysantana/gb
```

## Usage

Run benchmarks:

```bash
# Run benchmarks for last 6 commits (includes HEAD).
gb backfill HEAD~5
```

![Backfill Example](./imgs/backfill.png)

View benchmark notes for a specific commit:

```bash
gb show HEAD~3
```

![Show Example](./imgs/show.png)

View all benchmark notes for a specific commit (e.g. ones pushed from remote):

```bash
gb show HEAD~4 --all
```

(arm64 vs amd64 benchmarks shown below)
![Show All Example](./imgs/show-all.png)

Compare benchmarks between two commits:

```bash
gb compare HEAD~3 HEAD~1
# The default is HEAD~1 vs HEAD
gb compare
```

![Compare Example](./imgs/compare.png)

Sync notes with remote:

```bash
gb sync
```

![Sync Example](./imgs/sync.png)

## Manual

NAME:

-   gb - Go benchmark notes manager

USAGE:

-   gb [global options] [command [command options]]

COMMANDS:

-   Backfill benchmark notes with missing commits since a ref
-   Compare stored notes for two refs
-   Show stored note for a commit/ref
-   Sync benchmark notes with remote (push/fetch)

GLOBAL OPTIONS:

-   -v, --verbose (default: false)
-   --count int benchmark count (default: 10)
-   --benchtime string benchtime duration (e.g. 2s)
-   --bench string benchmark regex (default: ".")
-   --pkgs string comma-separated package list (default: "./...")
-   --notes-ref string override notes ref (will always be prefixed with refs/notes/gb/) (default: "601e17bd28b350c5/linux-amd64-go1.25.1")
-   --help, -h show help
