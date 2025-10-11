import { createResource } from "solid-js";
import { Route, Router } from "@solidjs/router";
import { loadManifest, type Manifest } from "./lib/data";
import Layout from "./components/Layout";
import Dashboard from "./components/Dashboard";

export default function App() {
    const [manifest] = createResource<Manifest>(loadManifest);
    
    return (
        <Layout manifest={manifest()}>
            <Router>
                <Route path="/" component={Dashboard} />
                <Route path="/:commit" component={Dashboard} />
                <Route path="/:commit/:benchmark" component={Dashboard} />
                <Route path="/:commit/:benchmark/:metric" component={Dashboard} />
            </Router>
        </Layout>
    );
}