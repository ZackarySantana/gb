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

export type Benchmark = BenchCase;

export const getBenchmark = query(async (name: string, commit: string) => {
    const res = await fetch(`/data/benchmarks/${name}/${commit}.json`);
    if (!res.ok) throw new Error(`benchmark load failed: ${res.status}`);
    return res.json() as Promise<Benchmark>;
}, "benchmark");
