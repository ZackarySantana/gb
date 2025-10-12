import { For, Suspense, type Component } from "solid-js";
import { createManifest } from "./primitives/createManifest";
import { createNote } from "./primitives/createNote";
import { Layout } from "./views/Layout";

const Dashboard: Component = () => {
    const manifest = createManifest();

    return (
        <Layout>
            <For each={manifest()?.benchmarks ?? []}>
                {(bench) => <div>Benchmark: {bench.name}</div>}
            </For>
            <For each={manifest()?.commits ?? []}>
                {(commit) => <Commit commit={commit} />}
            </For>
        </Layout>
    );
};

export function Card() {
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
                <a href="#" class="text-link hover:text-accent transition">
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
