import { createEffect, createMemo, createResource, createSignal, For, Show } from "solid-js";
import { useSearchParams, useParams } from "@solidjs/router";
import { loadManifest, loadNote, loadRun, noteToRunFile, type Manifest, type Note, type RunFile } from "../lib/data";
import RunSummary from "./RunSummary";
import Trend from "./Trend";

export default function Dashboard() {
    const [searchParams, setSearchParams] = useSearchParams();
    const params = useParams();
    
    // Load manifest
    const [manifest] = createResource<Manifest>(loadManifest);
    
    const initialCommit = (Array.isArray(params.commit) ? params.commit[0] : params.commit) || 
                         (Array.isArray(searchParams.commit) ? searchParams.commit[0] : searchParams.commit) || 
                         "";
    const initialBenchmark = (Array.isArray(params.benchmark) ? params.benchmark[0] : params.benchmark) || 
                            (Array.isArray(searchParams.benchmark) ? searchParams.benchmark[0] : searchParams.benchmark) || 
                            "";
    const initialMetric = (Array.isArray(params.metric) ? params.metric[0] : params.metric) || 
                         (Array.isArray(searchParams.metric) ? searchParams.metric[0] : searchParams.metric) || 
                         "";
    
    // URL-based state management - initialize directly from URL params
    const [commitId, setCommitId] = createSignal(initialCommit);
    const [benchName, setBenchName] = createSignal(initialBenchmark);
    const [metric, setMetric] = createSignal(initialMetric);
    
    // Update URL when state changes
    createEffect(() => {
        const urlParams: Record<string, string> = {};
        if (commitId() && commitId() !== "") urlParams.commit = commitId();
        if (benchName() && benchName() !== "") urlParams.benchmark = benchName();
        if (metric() && metric() !== "") urlParams.metric = metric();
        setSearchParams(urlParams, { scroll: false });
    });
    
    // No auto-selection - user must manually select
    
    // Get selected run path - only when a commit is actually selected
    const selectedRunPath = createMemo(() => {
        const m = manifest();
        const id = commitId();
        if (!m || !id || id === "") return undefined;
        return m.runs.find((r) => r.id === id)?.path;
    });
    
    // Load selected run file (handle both old and new formats)
    const [run] = createResource<
        RunFile | undefined,
        string | undefined
    >(selectedRunPath, async (p) => {
        if (!p) return undefined;
        
        try {
            // Try to load as new Note format first
            const note = await loadNote(p);
            return noteToRunFile(note);
        } catch (error) {
            // Fallback to old RunFile format
            return await loadRun(p);
        }
    });
    
    // No auto-selection - user must manually select
    
    // Build time series for trend chart
    const [series, setSeries] = createSignal<{ x: number[]; y: number[] }>({
        x: [],
        y: [],
    });
    
    createEffect(async () => {
        const m = manifest();
        const focus = benchName();
        const mtr = metric();
        if (!m || !focus || focus === "" || !mtr || mtr === "") return;
        
        // Get the last 10 runs in chronological order (oldest to newest)
        const last = m.runs.slice(-10);
        const xs: number[] = [];
        const ys: number[] = [];
        
        for (let i = 0; i < last.length; i++) {
            const r = last[i];
            let runFile: RunFile;
            let timestamp: number;
            
            try {
                // Try to load as new Note format first
                const note = await loadNote(r.path);
                runFile = noteToRunFile(note);
                // Use the actual timestamp from the note
                const parsedTimestamp = new Date(note.created_at).getTime() / 1000;
                if (!isNaN(parsedTimestamp)) {
                    timestamp = parsedTimestamp;
                } else {
                    // Fallback to date-based timestamp
                    timestamp = new Date(runFile.date + "T12:00:00Z").getTime() / 1000;
                }
            } catch (error) {
                // Fallback to old RunFile format
                runFile = await loadRun(r.path);
                // Use date-based timestamp for old format
                timestamp = new Date(runFile.date + "T12:00:00Z").getTime() / 1000;
            }
            
            const b = runFile.benchmarks.find(
                (b) => b.name === focus && b.metric === mtr
            );
            if (b) {
                xs.push(timestamp);
                ys.push(b.value);
            }
        }
        setSeries({ x: xs, y: ys });
    });
    
    // Computed options for filters
    const benchOptions = createMemo(() => {
        const r = run();
        if (!r) return [];
        const names = new Set(r.benchmarks.map((b) => b.name));
        return Array.from(names);
    });
    
    const metricOptions = createMemo(() => {
        const r = run();
        if (!r) return [];
        const mets = new Set(r.benchmarks.map((b) => b.metric));
        return Array.from(mets);
    });
    
    const commitOptions = createMemo(() => {
        const m = manifest();
        if (!m) return [];
        // Sort by date (newest first) and create better labels
        return m.runs
            .sort((a, b) => b.date.localeCompare(a.date)) // Sort by date descending (newest first)
            .map((r) => {
                // Parse date as local time to avoid timezone issues
                const [year, month, day] = r.date.split('-').map(Number);
                const date = new Date(year, month - 1, day);
                const formattedDate = date.toLocaleDateString('en-US', {
                    month: 'short',
                    day: 'numeric',
                    year: 'numeric'
                });
                return {
                    id: r.id,
                    label: `${r.id.slice(0, 7)} • ${formattedDate}`,
                };
            });
    });
    

    return (
        <div class="dashboard">
            {/* Commit View Section */}
            <div class="section">
                <div class="section-header">
                    <h2>Commit View</h2>
                    <div class="section-filter">
                        <label class="filter-label" for="commit-select">
                            Date
                        </label>
                        <select
                            id="commit-select"
                            value={commitId() || ""}
                            onChange={(e) => setCommitId(e.currentTarget.value)}
                            class="filter-select"
                        >
                            <option value="">Select commit...</option>
                            <Show when={commitOptions().length > 0} fallback={<option>Loading...</option>}>
                                <For each={commitOptions()}>
                                    {(opt) => (
                                        <option value={opt.id}>{opt.label}</option>
                                    )}
                                </For>
                            </Show>
                        </select>
                    </div>
                </div>
                <RunSummary run={run()} manifest={manifest()} />
            </div>
            
            {/* Benchmark View Section */}
            <div class="section">
                <div class="section-header">
                    <h2>Benchmark View</h2>
                    <div class="section-filters">
                        <div class="section-filter">
                            <label class="filter-label" for="bench-select">
                                Benchmark
                            </label>
                            <select
                                id="bench-select"
                                value={benchName() || ""}
                                onChange={(e) => setBenchName(e.currentTarget.value)}
                                class="filter-select"
                            >
                                <option value="">Select benchmark...</option>
                                <Show when={benchOptions().length > 0} fallback={<option>Loading...</option>}>
                                    <For each={benchOptions()}>
                                        {(name) => <option value={name}>{name}</option>}
                                    </For>
                                </Show>
                            </select>
                        </div>
                        <div class="section-filter">
                            <label class="filter-label" for="metric-select">
                                Metric
                            </label>
                            <select
                                id="metric-select"
                                value={metric() || ""}
                                onChange={(e) => setMetric(e.currentTarget.value)}
                                class="filter-select"
                            >
                                <option value="">Select metric...</option>
                                <Show when={metricOptions().length > 0} fallback={<option>Loading...</option>}>
                                    <For each={metricOptions()}>
                                        {(m) => <option value={m}>{m}</option>}
                                    </For>
                                </Show>
                            </select>
                        </div>
                    </div>
                </div>
                <Show when={series().x.length > 1}>
                    <Trend
                        title={`${benchName()} • ${metric()}`}
                        series={series()}
                    />
                </Show>
            </div>
        </div>
    );
}