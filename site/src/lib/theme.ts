// src/theme.ts
export const THEMES = [
    "Midnight",
    "Blueish",
    "Ocean",
    "Dracula",
    "Sunrise",
    "Inferno",
    "Mongo",
] as const;
export type ThemeName = (typeof THEMES)[number];
export const THEME_CHANGE_EVENT = "theme-change";
const KEY = "theme";

const isTheme = (x: string | null): x is ThemeName =>
    !!x && (THEMES as readonly string[]).includes(x);

function fromQuery(): ThemeName | null {
    try {
        const q = new URLSearchParams(window.location.search).get("theme");
        return isTheme(q) ? q : null;
    } catch {
        return null;
    }
}

function random(): ThemeName {
    return THEMES[Math.floor(Math.random() * THEMES.length)];
}

export function initTheme() {
    const chosen =
        fromQuery() ||
        (isTheme(localStorage.getItem(KEY))
            ? (localStorage.getItem(KEY) as ThemeName)
            : random());

    setTheme(chosen);
}

export function setTheme(name: ThemeName) {
    document.documentElement.setAttribute("data-theme", name);
    localStorage.setItem(KEY, name);

    // Custom event for listening to theme changes
    document.dispatchEvent(
        new CustomEvent(THEME_CHANGE_EVENT, { detail: name })
    );
}
