import { Accessor } from "solid-js";
import { getManifest, Manifest } from "../lib/data";
import { createAsync } from "@solidjs/router";

export function createManifest(): Accessor<Manifest | undefined> {
    return createAsync<Manifest>(() => getManifest());
}
