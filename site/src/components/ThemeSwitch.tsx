import { setTheme, THEMES, type ThemeName } from "../lib/theme";

export default function ThemeSwitcher(props: { class?: string }) {
    return (
        <div class={`flex items-center gap-2 ${props.class}`}>
            <select
                class="px-4 py-2 rounded border border-border bg-bg-elevated text-text-secondary text-fg outline-none"
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
