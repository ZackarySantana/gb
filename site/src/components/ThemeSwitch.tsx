import { setTheme, THEMES, type ThemeName } from "../lib/theme";
import Select from "./Select";

export default function ThemeSwitcher(props: { class?: string }) {
    return (
        <div class={`flex items-center gap-2 ${props.class}`}>
            <Select
                onInput={(e) =>
                    setTheme((e.target as HTMLSelectElement).value as ThemeName)
                }
                value={
                    document.documentElement.getAttribute("data-theme") || ""
                }
            >
                {THEMES.map((t) => (
                    <option value={t}>{t}</option>
                ))}
            </Select>
        </div>
    );
}
