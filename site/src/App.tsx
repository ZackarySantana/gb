import { createEffect, For, Suspense, type Component } from "solid-js";
import { createManifest } from "./primitives/createManifest";
import { createNote } from "./primitives/createCommit";
import ThemeSwitcher from "./components/ThemeSwitch";
import { A } from "@solidjs/router";
import { Manifest } from "./lib/data";

const App: Component = () => {
    const manifest = createManifest();

    return (
        <div>
            <Header manifest={manifest()} />
            <h1>Benchmark Manifest</h1>
            <Card />
            <ThemeSwitcher />
            <For each={manifest()?.commits ?? []}>
                {(commit) => <Commit commit={commit} />}
            </For>
        </div>
    );
};

function Header(props: { manifest: Manifest | undefined }) {
    // get generated at date
    // parse string to date
    const parsedDate = () =>
        new Date(props.manifest?.generated_at ?? Date.now());

    console.log("Parsed date:", parsedDate());

    return (
        <nav class="w-full sticky py-4 px-20 bg-bg-surface border-b border-border flex justify-between items-center gap-6">
            <div>
                <A
                    href="/"
                    class="text-2xl bg-gradient-to-r from-grad-from to-grad-to bg-clip-text text-transparent font-bold hover:bg-gradient-to-l block mb-1"
                >
                    go • benchmarks
                </A>
                <A
                    href="https://github.com/zackarysantana/gb"
                    class="text-xs text-text-secondary font-mono hover:text-accent-hover transition-all"
                >
                    {props.manifest?.module ?? "unknown module"}
                </A>
            </div>
            <ThemeSwitcher class="ml-auto" />
            <p class="text-xs text-text-secondary">{parsedDate().toString()}</p>
        </nav>
    );
}

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

    createEffect(() => {
        console.log("Loaded note:", note());
    });

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

export default App;
