import { JSX, Show } from "solid-js";
import type { Manifest } from "../lib/data";

interface LayoutProps {
    children: JSX.Element;
    manifest?: Manifest;
    onHomeClick?: () => void;
}

export default function Layout(props: LayoutProps) {
    return (
        <div class="layout">
            <header class="header">
                <div class="header-content">
                    <div class="header-left">
                        <button onClick={props.onHomeClick} class="title-link">
                            <h1 class="title">gb • benchmarks</h1>
                        </button>
                        <Show when={props.manifest}>
                            {(manifest) => (
                                <a 
                                    href={`https://${manifest().repo}`}
                                    target="_blank"
                                    rel="noreferrer"
                                    class="repo-link"
                                >
                                    {manifest().repo}
                                </a>
                            )}
                        </Show>
                    </div>
                    <div class="header-info">
                        <Show when={props.manifest} fallback={<span class="loading">Loading manifest…</span>}>
                            {(manifest) => (
                                <span class="last-updated">
                                    Updated {new Date(manifest().generatedAt).toLocaleString()}
                                </span>
                            )}
                        </Show>
                    </div>
                </div>
            </header>
            
            <main class="main">
                {props.children}
            </main>
        </div>
    );
}
