/* @refresh reload */
import "./index.css";
import { render } from "solid-js/web";
import { HashRouter, Route, Router } from "@solidjs/router";
import "solid-devtools";

import { initTheme } from "./lib/theme";
import Dashboard from "./Dashboard";
import NotFound from "./404";
import { Layout } from "./views/Layout";
import Benchmarks from "./Benchmarks";

const root = document.getElementById("root");

if (import.meta.env.DEV && !(root instanceof HTMLElement)) {
    throw new Error(
        "Root element not found. Did you forget to add it to your index.html? Or maybe the id attribute got misspelled?"
    );
}

initTheme();

render(
    () => (
        <HashRouter root={Layout}>
            <Route path="/" component={Dashboard} />
            <Route path="/:benchmark/commit/:commit" component={Benchmarks} />
            <Route path="*" component={NotFound} />
        </HashRouter>
    ),
    root!
);
