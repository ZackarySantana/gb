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
import { Benchmark, Stats, StatsKey } from "./lib/data";
import { createBenchmark } from "./primitives/createBenchmark";
import uPlot from "uplot";
import { createProperty } from "./primitives/createProperty";
import Select from "./components/Select";

import "uplot/dist/uPlot.min.css";

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

    let cursorInfo!: HTMLDivElement;
    let plotRef!: HTMLDivElement;
    let plot!: uPlot;
    let [toolTip, setToolTip] = createSignal<{
        x: number;
        y: number;
        commit: string;
    } | null>(null);

    onMount(() => {
        const accent = createProperty("--color-accent");
        const accentHover = createProperty("--color-accent-hover");
        const grid = createProperty("--color-bg-app");
        const text = createProperty("--color-text-secondary");

        plot = new uPlot(
            {
                width: plotRef.clientWidth,
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
                    lock: true,
                    focus: {
                        prox: 16,
                    },
                },
            },
            [],
            plotRef
        );

        let redraw = () => plot.redraw();

        document.addEventListener("theme-change", redraw);

        let ro = new ResizeObserver(() => {
            plot.setSize({
                width: plotRef.clientWidth,
                height: plot.height,
            });
            console.log("Resized plot", props.name);
        });
        ro.observe(plotRef);

        return () => {
            document.removeEventListener("theme-change", redraw);
            ro.disconnect();
            plot.destroy();
        };
    });

    createEffect(() => {
        const values = results()
            .map<[string, number | undefined]>((r) => [
                r[0],
                r[1].stats[yAxis()],
            ])
            .filter((v) => v[1] !== undefined && !isNaN(v[1]!)) as [
            string,
            number
        ][];

        const statInfo = Stats[yAxis()];

        const xValues = values.map((_, i) => i);
        const yValues = values.map((v) => statInfo.conversion(v[1]));
        const commitLabels = values.map(([commit]) => commit.slice(0, 5));

        plot.axes[0].values = (_, splits) => {
            return splits.map((i) => {
                const idx = Math.round(i);
                return commitLabels[idx] ?? "";
            });
        };
        plot.axes[0].label = "Commit";
        plot.axes[1].label = statInfo.label;

        plot.hooks.setCursor = [
            (u) => {
                const i = u.cursor.idx;
                console.log("cursor idx: ", i);
                if (i != null && i >= 0 && i < commitLabels.length) {
                    const commit = commitLabels[i];
                    console.log("hovering over: ", commit);
                    // use commit (tooltip, link, etc.)
                    setToolTip({
                        x: u.cursor.left ?? 0,
                        y: u.cursor.top ?? 0,
                        commit: commit,
                    });
                }
            },
        ];

        plot.setData([xValues, yValues]);
        console.log("Updated plot data", props.name);
    });

    createEffect(() => {
        console.log("yaxis changed", yAxis());
    });

    return (
        <div class="bg-bg-surface text-text-primary border border-border rounded-lg p-5">
            <div class="flex gap-5 items-center">
                <h1>{props.name}</h1>
                <Select
                    class="text-sm"
                    onInput={(e) =>
                        setYAxis(
                            (e.target as HTMLSelectElement).value as StatsKey
                        )
                    }
                >
                    <For each={Object.keys(Stats) as StatsKey[]}>
                        {(key) => (
                            <option value={key} selected={key === yAxis()}>
                                {Stats[key].label}
                            </option>
                        )}
                    </For>
                </Select>
            </div>
            <div class="relative">
                <div ref={plotRef} />

                <div
                    ref={cursorInfo}
                    class="absolute"
                    style={{
                        display: toolTip() ? "block" : "none",
                        left: toolTip() ? `${toolTip()!.x + 15}px` : "0px",
                        top: toolTip() ? `${toolTip()!.y + 30}px` : "0px",
                    }}
                >
                    {toolTip() && (
                        <div class="bg-bg-app text-text-primary border border-border rounded-md p-2 shadow-lg whitespace-nowrap">
                            <div>
                                <strong>Commit:</strong> {toolTip()!.commit}
                            </div>
                            <div>
                                <a
                                    href={`https://github.com/commit/${
                                        toolTip()!.commit
                                    }`}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                >
                                    View Commit
                                </a>
                            </div>
                        </div>
                    )}
                </div>
            </div>
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
