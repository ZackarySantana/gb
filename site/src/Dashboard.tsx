import {
    Accessor,
    createEffect,
    createMemo,
    createSignal,
    For,
    onMount,
    Show,
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
import GitHash from "./components/GitHash";
import { A } from "@solidjs/router";

type CommitInfo = {
    hash: string;
    commitTitle: string;
    author: string;
    date: string;
};

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
    let [toolTip, setToolTip] = createSignal<
        | ({
              x: number;
              y: number;
          } & CommitInfo)
        | null
    >(null);

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

        let ro = new ResizeObserver(() =>
            plot.setSize({ width: plotRef.clientWidth, height: plot.height })
        );
        ro.observe(plotRef);

        return () => {
            document.removeEventListener("theme-change", redraw);
            ro.disconnect();
            plot.destroy();
        };
    });

    createEffect(() => {
        const values = results()
            .map<[CommitInfo, number | undefined]>((r) => [
                {
                    hash: r[0],
                    commitTitle: r[1].commitTitle,
                    author: r[1].author,
                    date: r[1].date,
                },
                r[1].stats[yAxis()],
            ])
            .filter((v) => v[1] !== undefined && !isNaN(v[1]!)) as [
            CommitInfo,
            number
        ][];

        const statInfo = Stats[yAxis()];

        const xValues = values.map((_, i) => i);
        const yValues = values.map((v) => statInfo.conversion(v[1]));
        const commitLabels = values.map(([commit]) => commit.hash.slice(0, 5));

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
                if (i != null && i >= 0 && i < commitLabels.length) {
                    setToolTip({
                        x: u.cursor.left ?? 0,
                        y: u.cursor.top ?? 0,
                        hash: commitLabels[i],
                        author: values[i][0].author,
                        commitTitle: values[i][0].commitTitle,
                        date: values[i][0].date,
                    });
                } else {
                    setToolTip(null);
                }
            },
        ];

        plot.setData([xValues, yValues]);
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
                        left: toolTip()
                            ? `${
                                  toolTip()!.x +
                                  15 -
                                  (toolTip()!.x / plotRef.clientWidth) * 300
                              }px`
                            : "0px",
                        top: toolTip() ? `${toolTip()!.y + 30}px` : "0px",
                    }}
                >
                    <Show when={toolTip() != null}>
                        <div class="bg-bg-app text-text-primary border border-border rounded-md p-4 shadow-lg max-w-[350px]">
                            <div class="flex items-center gap-2">
                                <GitHash hash={toolTip()!.hash} />
                                <p class="whitespace-normal break-words line-clamp-1">
                                    {toolTip()!.commitTitle}
                                </p>
                            </div>
                            <div class="mt-2">
                                <A
                                    href={`/${props.name}/commit/${
                                        toolTip()!.hash
                                    }`}
                                    class="text-text-link hover:text-accent transition"
                                >
                                    Benchmarks
                                </A>
                                <p class="text-sm text-text-secondary">
                                    Date: {toolTip()!.date}
                                </p>
                                <p class="text-sm text-text-secondary">
                                    Author:{" "}
                                    <A
                                        href={`https://github.com/${
                                            toolTip()!.author
                                        }`}
                                        class="text-text-link hover:text-accent italic transition"
                                    >
                                        {toolTip()!.author}
                                    </A>
                                </p>
                            </div>
                        </div>
                    </Show>
                </div>
            </div>
        </div>
    );
}

export default Dashboard;
