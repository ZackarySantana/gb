import { createEffect, For, Show, createSignal } from "solid-js";
import type { RunFile, Manifest } from "../lib/data";

export default function RunSummary(props: { run?: RunFile; manifest?: Manifest }) {
    const [showMoreInfo, setShowMoreInfo] = createSignal(false);

    return (
        <div class="card">
            <Show when={props.run} fallback={<div class="loading">Loading latest run…</div>}>
                {(run) => (
                    <>
                        <div class="run-header">
                            <div class="commit-info">
                                <a
                                    href={
                                        props.manifest?.repo
                                            ? `https://${props.manifest.repo}/commit/${run().commit.hash}`
                                            : "#"
                                    }
                                    target="_blank"
                                    rel="noreferrer"
                                    class="commit-link"
                                >
                                    {run().commit.hash.slice(0, 7)}
                                </a>
                                <span class="created-time">
                                    {new Date(run().created_at).toLocaleString('en-US', {
                                        month: 'short',
                                        day: 'numeric',
                                        year: 'numeric',
                                        hour: 'numeric',
                                        minute: '2-digit',
                                        hour12: true
                                    })}
                                </span>
                            </div>
                            <button 
                                class="more-info-btn"
                                onClick={() => setShowMoreInfo(!showMoreInfo())}
                            >
                                {showMoreInfo() ? 'Less info' : 'More info'}
                            </button>
                        </div>

                        <Show when={showMoreInfo()}>
                            <div class="more-info">
                                <div class="info-section">
                                    <h4>Environment</h4>
                                    <div class="info-grid">
                                        <div class="info-item">
                                            <span class="info-label">Go Version:</span>
                                            <span class="info-value">{run().env.go_version}</span>
                                        </div>
                                        <div class="info-item">
                                            <span class="info-label">OS/Arch:</span>
                                            <span class="info-value">{run().env.goos}/{run().env.goarch}</span>
                                        </div>
                                        <div class="info-item">
                                            <span class="info-label">Max Procs:</span>
                                            <span class="info-value">{run().env.gomaxprocs}</span>
                                        </div>
                                        <div class="info-item">
                                            <span class="info-label">Host:</span>
                                            <span class="info-value">{run().env.host}</span>
                                        </div>
                                        <div class="info-item">
                                            <span class="info-label">CPUs:</span>
                                            <span class="info-value">{run().env.cpus}</span>
                                        </div>
                                    </div>
                                </div>
                                <div class="info-section">
                                    <h4>Benchmark Arguments</h4>
                                    <div class="bench-args">
                                        {run().bench_args.join(' ')}
                                    </div>
                                </div>
                            </div>
                        </Show>

                        <div class="table-container">
                            <table>
                                <thead>
                                    <tr>
                                        <th>Benchmark</th>
                                        <th>Metric</th>
                                        <th>Value</th>
                                        <th>Performance</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <For each={run().benchmarks.slice(0, 25)}>
                                        {(b) => {
                                            const delta =
                                                run().compare?.deltas.find(
                                                    (d) =>
                                                        d.name === b.name &&
                                                        d.metric === b.metric
                                                )?.deltaPct ?? null;
                                            
                                            const deltaClass = delta === null 
                                                ? "neutral" 
                                                : delta < 0 
                                                    ? "negative" 
                                                    : "positive";
                                            
                                            const formatValue = (value: number, metric: string) => {
                                                if (metric === "ns/op") {
                                                    return value < 1000 ? `${value.toFixed(1)} ns` : `${(value / 1000).toFixed(2)} μs`;
                                                } else if (metric === "B/op") {
                                                    return value < 1024 ? `${value} B` : `${(value / 1024).toFixed(1)} KB`;
                                                } else if (metric === "allocs/op") {
                                                    return `${value} allocs`;
                                                }
                                                return value.toString();
                                            };
                                            
                                            return (
                                                <tr>
                                                    <td>
                                                        <span class="benchmark-name">{b.name}</span>
                                                    </td>
                                                    <td>
                                                        <span class="metric-name">{b.metric}</span>
                                                    </td>
                                                    <td>
                                                        <span class="value">{formatValue(b.value, b.metric)}</span>
                                                    </td>
                                                    <td>
                                                        <Show
                                                            when={delta !== null}
                                                            fallback={
                                                                <span class="delta neutral">—</span>
                                                            }
                                                        >
                                                            <span class={`delta ${deltaClass}`}>
                                                                {delta! < 0 ? "↓" : "↑"} {Math.abs(delta!).toFixed(1)}%
                                                            </span>
                                                        </Show>
                                                    </td>
                                                </tr>
                                            );
                                        }}
                                    </For>
                                </tbody>
                            </table>
                        </div>
                    </>
                )}
            </Show>
        </div>
    );
}