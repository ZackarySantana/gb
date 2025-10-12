import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import solidPlugin from "vite-plugin-solid";
import devtools from "solid-devtools/vite";

export default defineConfig({
    plugins: [devtools(), solidPlugin(), tailwindcss()],
    server: {
        port: 3000,
    },
    build: {
        target: "esnext",
    },
    // Cache headers on preview.
    preview: {
        headers: {
            "Cache-Control": "public, max-age=31536000, immutable",
        },
    },
});
