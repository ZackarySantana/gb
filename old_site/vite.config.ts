import { defineConfig } from "vite";
import solid from "vite-plugin-solid";

export default defineConfig({
    plugins: [solid()],
    build: {
        target: "esnext",
        outDir: "dist",
        sourcemap: true,
    },
    server: {
        port: 5173,
    },
});
