import { Accessor } from "solid-js";
import { Benchmark, getBenchmark, Note } from "../lib/data";
import { createAsync } from "@solidjs/router";

export function createBenchmark(
    name: string,
    commit: string
): Accessor<Benchmark | undefined> {
    return createAsync(() => getBenchmark(name, commit));
}
