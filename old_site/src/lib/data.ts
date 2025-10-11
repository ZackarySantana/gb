export type Manifest = {
    generatedAt: string;
    defaultBaseline?: string;
    repo: string;
    package: string;
    runs: { id: string; date: string; path: string; count: number }[];
};

// Note represents the benchmark note stored as a git note
export type Note = {
    schema: number;
    commit: string;
    created_at: string; // ISO timestamp
    env: EnvInfo;
    bench_args: string[];
    parsed: BenchReport;
    raw: string;
};

// EnvInfo describes the Go runtime and host environment at benchmark time
export type EnvInfo = {
    go_version: string;
    goos: string;
    goarch: string;
    gomaxprocs: number;
    host: string;
    cpus: number;
};

// BenchReport corresponds to a single `go test -bench` output file
export type BenchReport = {
    goos?: string;
    goarch?: string;
    pkg?: string;
    cpu?: string;
    pass: boolean;
    duration: string; // e.g., "12.916s"
    benches: BenchCase[];
};

// BenchCase aggregates all repeated lines for one benchmark name
export type BenchCase = {
    name: string; // e.g., "BenchmarkConcatJoin-24"
    samples: BenchSample[];
    stats: BenchStats;
};

// BenchSample is a single printed benchmark line (one repetition)
export type BenchSample = {
    iterations: number;
    ns_per_op: number;
    bytes_per_op?: number;
    allocs_per_op?: number;
};

// BenchStats is a compact summary over all Samples for a BenchCase
export type BenchStats = {
    count: number;
    // Time
    ns_per_op_mean: number;
    ns_per_op_median: number;
    ns_per_op_min: number;
    ns_per_op_max: number;
    // Memory & allocs
    bytes_per_op_mean?: number;
    bytes_per_op_median?: number;
    allocs_per_op_mean?: number;
    allocs_per_op_median?: number;
};

// Legacy RunFile type for backward compatibility (will be removed)
export type RunFile = {
    date: string; // YYYY-MM-DD
    created_at: string; // ISO timestamp
    commit: { hash: string; title?: string };
    go: { version: string; os: string; arch: string };
    env: EnvInfo;
    bench_args: string[];
    benchmarks: {
        name: string;
        metric: string; // "ns/op", "B/op", etc.
        value: number;
        extra?: Record<string, number>;
        labels?: Record<string, string>;
    }[];
    compare?: {
        baseline: string;
        deltas: { name: string; metric: string; deltaPct: number }[];
    };
};

export async function loadManifest(): Promise<Manifest> {
    const res = await fetch("./data/manifest.json", { cache: "no-cache" });
    if (!res.ok) throw new Error(`manifest load failed: ${res.status}`);
    return res.json();
}

export async function loadRun(path: string): Promise<RunFile> {
    const res = await fetch(`./data/${path}`, { cache: "force-cache" });
    if (!res.ok) throw new Error(`run load failed: ${res.status}`);
    console.log("Loaded run from", path);
    return res.json();
}

export async function loadNote(path: string): Promise<Note> {
    const res = await fetch(`./data/${path}`, { cache: "force-cache" });
    if (!res.ok) throw new Error(`note load failed: ${res.status}`);
    console.log("Loaded note from", path);
    return res.json();
}

// Helper function to convert Note to RunFile for backward compatibility
export function noteToRunFile(note: Note): RunFile {
    // Parse the date more safely
    let date: string;
    try {
        const parsedDate = new Date(note.created_at);
        if (isNaN(parsedDate.getTime())) {
            // Fallback to a default date if parsing fails
            date = "2025-10-06";
        } else {
            date = parsedDate.toISOString().split('T')[0];
        }
    } catch (error) {
        console.warn("Failed to parse date:", note.created_at, error);
        date = "2025-10-06";
    }
    
    // Handle both old and new commit formats
    let commitHash: string;
    if (typeof note.commit === 'string') {
        commitHash = note.commit.slice(0, 7);
    } else if (note.commit && typeof note.commit === 'object' && 'hash' in note.commit) {
        // Old format: { hash: "abc123", title: "...", repo: "..." }
        commitHash = (note.commit as any).hash.slice(0, 7);
    } else {
        // Fallback
        commitHash = "unknown";
    }
    
    return {
        date,
        created_at: note.created_at,
        commit: { 
            hash: commitHash, 
            title: `commit ${commitHash}` // We don't have commit titles in notes
        },
        go: {
            version: note.env.go_version,
            os: note.env.goos,
            arch: note.env.goarch
        },
        env: note.env,
        bench_args: note.bench_args,
        benchmarks: note.parsed.benches.map(bench => ({
            name: bench.name,
            metric: "ns/op",
            value: bench.stats.ns_per_op_mean
        })),
        compare: undefined // We'll add comparison logic later
    };
}
