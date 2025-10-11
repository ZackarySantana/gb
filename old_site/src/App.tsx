import { createResource } from "solid-js";
import { Route, Router, useNavigate } from "@solidjs/router";
import { loadManifest, type Manifest } from "./lib/data";
import Layout from "./components/Layout";
import Dashboard from "./components/Dashboard";

function AppContent() {
    const [manifest] = createResource<Manifest>(loadManifest);
    const navigate = useNavigate();
    
    const handleHomeClick = () => {
        // Navigate to home and clear URL parameters
        navigate("/", { replace: true });
    };
    
    return (
        <Layout manifest={manifest()} onHomeClick={handleHomeClick}>
            <Route path="/" component={Dashboard} />
            <Route path="/:commit" component={Dashboard} />
            <Route path="/:commit/:benchmark" component={Dashboard} />
            <Route path="/:commit/:benchmark/:metric" component={Dashboard} />
        </Layout>
    );
}

export default function App() {
    return (
        <Router>
            <AppContent />
        </Router>
    );
}