import { colors, getRandomFunkyTheme } from "../lib/color_themes";

import { setTheme, THEMES, type ThemeName } from "../lib/theme";

export default function ThemeSwitcher() {
    return (
        <div class="flex items-center gap-2">
            <label class="text-muted">Theme</label>
            <select
                class="px-2 py-1 rounded border border-border bg-surface text-fg"
                onInput={(e) =>
                    setTheme((e.target as HTMLSelectElement).value as ThemeName)
                }
            >
                {THEMES.map((t) => (
                    <option value={t}>{t}</option>
                ))}
            </select>
        </div>
    );
}
