import {
    Accessor,
    createEffect,
    createMemo,
    createSignal,
    For,
    onMount,
    Suspense,
    type Component,
} from "solid-js";
import { createManifest } from "./primitives/createManifest";
import { createNote } from "./primitives/createNote";
import { Benchmark, StatsKey } from "./lib/data";
import { createBenchmark } from "./primitives/createBenchmark";
import { A } from "@solidjs/router";
import GitHash from "./components/GitHash";
import uPlot, { AlignedData } from "uplot";
import { createProperty } from "./primitives/createProperty";

const Dashboard: Component = () => {
    const manifest = createManifest();

    return (
        <div class="py-8">
            <div class="flex gap-5 flex-col">
                <For each={manifest()?.benchmarks ?? []}>
                    {(bench) => (
                        <BenchmarkCommit
                            name={bench.name}
                            commits={bench.commits}
                        />
                    )}
                </For>
            </div>
            <For each={manifest()?.commits ?? []}>
                {(commit) => <Commit commit={commit} />}
            </For>
        </div>
    );
};

function BenchmarkCommit(props: { name: string; commits: string[] }) {
    const rawResults: [string, Accessor<Benchmark | undefined>][] =
        props.commits.map((c) => [c, createBenchmark(props.name, c)]);

    const [yAxis, setYAxis] = createSignal<StatsKey>("ns_per_op_median");

    const results: Accessor<[string, Benchmark][]> = createMemo(() =>
        rawResults
            .map((v) => [v[0], v[1]()] as [string, Benchmark])
            .filter((v) => v[1] !== undefined)
    );

    let el!: HTMLDivElement;
    let plot!: uPlot;

    onMount(() => {
        const accent = createProperty("--color-accent");
        const accentHover = createProperty("--color-accent-hover");
        const grid = createProperty("--color-bg-app");
        const text = createProperty("--color-text-secondary");

        plot = new uPlot(
            {
                title: props.name,
                width: el.clientWidth,
                height: 320,
                pxAlign: 0,
                select: {
                    show: false,
                    left: 0,
                    top: 0,
                    width: 0,
                    height: 0,
                },
                scales: {
                    x: { distr: 2 },
                    y: { auto: true },
                },
                series: [
                    {},
                    {
                        label: "value",
                        width: 2,
                        stroke: accent,
                        points: {
                            show: true,
                            size: 6,
                            width: 2,
                            stroke: accentHover,
                            fill: accent,
                        },
                    },
                ],
                axes: [
                    {
                        grid: {
                            stroke: grid,
                            width: 1,
                        },
                        ticks: {
                            stroke: grid,
                            width: 1,
                        },
                        font: "12px system-ui, -apple-system, sans-serif",
                        stroke: text,
                    },
                    {
                        grid: {
                            stroke: grid,
                            width: 1,
                        },
                        ticks: {
                            stroke: grid,
                            width: 1,
                        },
                        font: "12px system-ui, -apple-system, sans-serif",
                        stroke: text,
                    },
                ],
                legend: {
                    show: false,
                },
                cursor: {
                    drag: { x: true, y: true },
                    points: {
                        show: true,
                        size: 8,
                        width: 2,
                        stroke: "#ffffff",
                        fill: "#3b82f6",
                    },
                    show: true,
                    lock: false,
                    focus: {
                        prox: 16,
                    },
                },
            },
            [],
            el
        );

        document.addEventListener("theme-change", () => {
            plot.redraw();
        });
    });

    createEffect(() => {
        const values = results()
            .map<[string, number | undefined]>((r) => [
                r[0],
                r[1].stats[yAxis()],
            ])
            .filter((v) => v[1] !== undefined && !isNaN(v[1]!));

        const xValues = values.map((_, i) => i);
        const yValues = values.map((v) => v[1]);
        const commitLabels = values.map(([commit]) => commit.slice(0, 5));

        plot.axes[0].values = (_, splits) => {
            return splits.map((i) => {
                const idx = Math.round(i);
                return commitLabels[idx] ?? "";
            });
        };

        plot.hooks.setCursor = [
            (u) => {
                const i = u.cursor.idx;
                console.log("cursor idx: ", i);
                if (i != null && i >= 0 && i < commitLabels.length) {
                    const commit = commitLabels[i];
                    console.log("hovering over: ", commit);
                    // use commit (tooltip, link, etc.)
                }
            },
        ];

        plot.setData([xValues, yValues]);
    });

    return (
        <div class="bg-bg-surface text-text-primary border border-border rounded-lg p-5">
            <div ref={el} />
        </div>
    );
}

function Card() {
    return (
        <div class="rounded-xl p-6 bg-bg-surface text-text-primary shadow-sm">
            <h2 class="text-xl font-semibold">Hello!</h2>
            <p class="text-secondary mt-1">Welcome to my site.</p>

            <div class="mt-4 flex gap-3">
                <button class="px-4 py-2 rounded-md bg-btn-primary text-on-btn-primary hover:bg-btn-primary-hover transition">
                    Action
                </button>
                <button class="px-4 py-2 rounded-md border border-btn-outline-border text-btn-outline-fg hover:bg-[var(--btn-outline-hover)] transition">
                    Bordered
                </button>
            </div>

            <div class="mt-6 h-1 w-full rounded-full bg-gradient-to-r from-grad-from to-grad-to" />

            <div class="mt-6 rounded-lg p-4 bg-elevated border border-border">
                <a href="#" class="text-text-link hover:text-accent transition">
                    @zack
                </a>
            </div>
        </div>
    );
}

const Commit: Component<{ commit: { hash: string; note_refs: string[] } }> = (
    props
) => {
    return (
        <div>
            <h2>Commit: {props.commit.hash}</h2>
            <ul>
                <For each={props.commit.note_refs}>
                    {(note) => (
                        <Note commit={props.commit.hash} noteRef={note} />
                    )}
                </For>
            </ul>
        </div>
    );
};

const Note: Component<{ commit: string; noteRef: string }> = (props) => {
    const note = createNote(props.commit, props.noteRef);

    return (
        <li>
            <Suspense fallback={<span>Loading note {props.noteRef}...</span>}>
                <div>
                    <strong>Note: {props.noteRef}</strong>
                    <div>Created at: {note()?.created_at}</div>
                    <div>Go version: {note()?.env.go_version}</div>
                    <div>Host: {note()?.env.host}</div>
                    <div>Benches:</div>
                    <ul>
                        <For each={note()?.parsed.benches ?? []}>
                            {(bench) => (
                                <li>
                                    <strong>{bench.name}</strong>
                                    <ul>
                                        <li>
                                            Count: {bench.stats.count}, Ns/op
                                            Mean:{" "}
                                            {bench.stats.ns_per_op_mean.toFixed(
                                                2
                                            )}
                                            , Ns/op Median:{" "}
                                            {bench.stats.ns_per_op_median.toFixed(
                                                2
                                            )}
                                            , Ns/op Min:{" "}
                                            {bench.stats.ns_per_op_min.toFixed(
                                                2
                                            )}
                                            , Ns/op Max:{" "}
                                            {bench.stats.ns_per_op_max.toFixed(
                                                2
                                            )}
                                            {bench.stats.bytes_per_op_mean !==
                                                undefined && (
                                                <>
                                                    , Bytes/op Mean:{" "}
                                                    {bench.stats.bytes_per_op_mean.toFixed(
                                                        2
                                                    )}
                                                    , Bytes/op Median:{" "}
                                                    {/* {bench.stats.bytes_per_op_median.toFixed(
                                                                2
                                                            )}
                                                            , Bytes/op Min:{" "}
                                                            {bench.stats.bytes_per_op_min.toFixed(
                                                                2
                                                            )}
                                                            , Bytes/op Max:{" "}
                                                            {bench.stats.bytes_per_op_max.toFixed(
                                                                2
                                                            )} */}
                                                </>
                                            )}
                                        </li>
                                    </ul>
                                </li>
                            )}
                        </For>
                    </ul>
                </div>
            </Suspense>
        </li>
    );
};

export default Dashboard;
