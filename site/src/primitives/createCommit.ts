import { Accessor } from "solid-js";
import { getNote, Note } from "../lib/data";
import { createAsync } from "@solidjs/router";

export function createNote(
    commit: string,
    ref: string
): Accessor<Note | undefined> {
    return createAsync<Note>(() => getNote(commit, ref));
}
