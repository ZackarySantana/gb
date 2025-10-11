import { query } from "@solidjs/router";

export type Manifest = {
    generated_at: string;
    latest_commit: string;
    module: string;
    commits: {
        hash: string;
        note_refs: string[];
    }[];
};

export const getManifest = query(async () => {
    const res = await fetch("./data/manifest.json");
    if (!res.ok) throw new Error(`manifest load failed: ${res.status}`);
    return res.json() as Promise<Manifest>;
}, "manifest");

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
        benches: {
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
        }[];
    };
};

export const getNote = query(async (commit: string, ref: string) => {
    const res = await fetch(`./data/${commit}/${ref}.json`);
    if (!res.ok) throw new Error(`note load failed: ${res.status}`);
    return res.json() as Promise<Note>;
}, "note");
