import { For } from "solid-js";

interface FilterOption {
    id: string;
    label: string;
}

interface FiltersProps {
    commitId: string;
    setCommitId: (id: string) => void;
    benchName: string;
    setBenchName: (name: string) => void;
    metric: string;
    setMetric: (metric: string) => void;
    commitOptions: FilterOption[];
    benchOptions: string[];
    metricOptions: string[];
    onReload: () => void;
}

export default function Filters(props: FiltersProps) {
    return (
        <div class="filters-card">
            <div class="filters-grid">
                {/* Commit selector */}
                <div class="filter-group">
                    <label class="filter-label" for="commit-select">
                        Commit / run
                    </label>
                    <select
                        id="commit-select"
                        value={props.commitId}
                        onChange={(e) => props.setCommitId(e.currentTarget.value)}
                        class="filter-select"
                    >
                        <For each={props.commitOptions}>
                            {(opt) => (
                                <option value={opt.id}>{opt.label}</option>
                            )}
                        </For>
                    </select>
                </div>

                {/* Benchmark selector */}
                <div class="filter-group">
                    <label class="filter-label" for="bench-select">
                        Benchmark
                    </label>
                    <select
                        id="bench-select"
                        value={props.benchName}
                        onChange={(e) => props.setBenchName(e.currentTarget.value)}
                        class="filter-select"
                    >
                        <For each={props.benchOptions}>
                            {(name) => <option value={name}>{name}</option>}
                        </For>
                    </select>
                </div>

                {/* Metric selector */}
                <div class="filter-group">
                    <label class="filter-label" for="metric-select">
                        Metric
                    </label>
                    <select
                        id="metric-select"
                        value={props.metric}
                        onChange={(e) => props.setMetric(e.currentTarget.value)}
                        class="filter-select"
                    >
                        <For each={props.metricOptions}>
                            {(m) => <option value={m}>{m}</option>}
                        </For>
                    </select>
                </div>

                {/* Reload button */}
                <div class="filter-group">
                    <label class="filter-label">&nbsp;</label>
                    <button
                        onClick={props.onReload}
                        title="Reload selected run"
                        class="reload-button"
                    >
                        Reload
                    </button>
                </div>
            </div>
        </div>
    );
}
