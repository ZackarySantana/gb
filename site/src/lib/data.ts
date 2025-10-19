import { query } from "@solidjs/router";

export type Manifest = {
    generated_at: string;
    latest_commit: string;
    module: string;
    commits: {
        hash: string;
        note_refs: string[];
    }[];
    benchmarks: {
        name: string;
        commits: string[];
    }[];
};

export const getManifest = query(async () => {
    const res = await fetch("/data/manifest.json", { cache: "force-cache" });
    if (!res.ok) throw new Error(`manifest load failed: ${res.status}`);
    return res.json() as Promise<Manifest>;
}, "manifest");

type BenchCase = {
    name: string;
    samples: {
        iterations: number;
        ns_per_op: number;
        bytes_per_op?: number;
        allocs_per_op?: number;
    }[];
    stats: {
        count: number;
        // Time
        ns_per_op_mean: number;
        ns_per_op_median: number;
        ns_per_op_min: number;
        ns_per_op_max: number;
        // Memory & allocs (use -benchmem to populate)
        bytes_per_op_mean?: number;
        bytes_per_op_median?: number;
        bytes_per_op_min?: number;
        bytes_per_op_max?: number;
        allocs_per_op_mean?: number;
        allocs_per_op_median?: number;
        allocs_per_op_min?: number;
        allocs_per_op_max?: number;
    };
};

export type StatsKey = keyof BenchCase["stats"];

export type StatInfo = {
    label: string;
    conversion: (v: number) => number;
};

export const Stats: Record<StatsKey, StatInfo> = {
    count: { label: "Samples", conversion: (v) => v },
    ns_per_op_mean: {
        label: "Mean Time (ms)",
        conversion: (v) => v / 1_000_000,
    },
    ns_per_op_median: {
        label: "Median Time (ms)",
        conversion: (v) => v / 1_000_000,
    },
    ns_per_op_min: {
        label: "Min Time (ms)",
        conversion: (v) => v / 1_000_000,
    },
    ns_per_op_max: {
        label: "Max Time (ms)",
        conversion: (v) => v / 1_000_000,
    },
    bytes_per_op_mean: {
        label: "Mean Bytes",
        conversion: (v) => v,
    },
    bytes_per_op_median: {
        label: "Median Bytes",
        conversion: (v) => v,
    },
    bytes_per_op_min: {
        label: "Min Bytes",
        conversion: (v) => v,
    },
    bytes_per_op_max: {
        label: "Max Bytes",
        conversion: (v) => v,
    },
    allocs_per_op_mean: {
        label: "Mean Allocs",
        conversion: (v) => v,
    },
    allocs_per_op_median: {
        label: "Median Allocs",
        conversion: (v) => v,
    },
    allocs_per_op_min: {
        label: "Min Allocs",
        conversion: (v) => v,
    },
    allocs_per_op_max: {
        label: "Max Allocs",
        conversion: (v) => v,
    },
};

export type Note = {
    schema: number;
    commit: string;
    created_at: string;
    env: {
        go_version: string;
        goos: string;
        goarch: string;
        gomaxprocs: number;
        host: string;
        cpus: number;
    };
    bench_args: string[];
    parsed: {
        // TODO: Should we include all the other fields? Look in note.go
        benches: BenchCase[];
    };
};

export const getNote = query(async (commit: string, ref: string) => {
    const res = await fetch(`/data/commits/${commit}/${ref}.json`, {
        cache: "force-cache",
    });
    if (!res.ok) throw new Error(`note load failed: ${res.status}`);
    return res.json() as Promise<Note>;
}, "note");

export type Benchmark = BenchCase & {
    commitTitle: string;
    author: string;
    date: Date;
};

export const getBenchmark = query(async (name: string, commit: string) => {
    const res = await fetch(`/data/benchmarks/${name}/${commit}.json`);
    if (!res.ok) throw new Error(`benchmark load failed: ${res.status}`);
    return res.json() as Promise<Benchmark>;
}, "benchmark");
