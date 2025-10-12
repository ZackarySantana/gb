import { A } from "@solidjs/router";
import { createManifest } from "../primitives/createManifest";

export default function GitHash(props: { hash: string; class?: string }) {
    const manifest = createManifest();

    return (
        <A
            href={`https://${manifest()?.module}/commit/${props.hash}`}
            class="bg-bg-elevated text-text-link hover:text-accent py-2 px-2 rounded-lg ml-auto text-sm font-mono transition"
        >
            {props.hash.slice(0, 6)}
        </A>
    );
}
