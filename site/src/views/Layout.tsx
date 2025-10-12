import { Component, JSX } from "solid-js";
import { createManifest } from "../primitives/createManifest";
import { A } from "@solidjs/router";
import ThemeSwitcher from "../components/ThemeSwitch";

export const Layout: Component<{ children: JSX.Element }> = (props) => {
    return (
        <>
            <Header />
            <div class="px-16">{props.children}</div>
        </>
    );
};

function Header() {
    const manifest = createManifest();

    const parsedDate = () => new Date(manifest()?.generated_at ?? Date.now());

    return (
        <nav class="w-full sticky py-4 px-16 bg-bg-surface border-b border-border flex justify-between items-center gap-6 top-0 z-10">
            <div>
                <A
                    href="/"
                    class="text-2xl bg-gradient-to-r from-grad-from to-grad-to bg-clip-text text-transparent font-bold hover:bg-gradient-to-l block mb-1"
                >
                    go • benchmarks
                </A>
                <A
                    href={`https://${manifest()?.module ?? "github.com"}`}
                    class="text-xs text-text-secondary font-mono hover:text-accent-hover transition-all"
                >
                    {manifest()
                        ? manifest()?.module ?? "unknown module"
                        : "loading..."}
                </A>
            </div>
            <ThemeSwitcher class="ml-auto text-xs" />
            <p class="text-xs text-text-secondary">{parsedDate().toString()}</p>
        </nav>
    );
}
