import { setTheme, THEMES, type ThemeName } from "../lib/theme";

export default function ThemeSwitcher(props: { class?: string }) {
    return (
        <div class={`flex items-center gap-2 ${props.class}`}>
            <select
                class="px-6 py-2 rounded border border-border bg-bg-app text-text-primary text-fg outline-none"
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
