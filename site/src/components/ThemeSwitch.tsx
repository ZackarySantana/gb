import { setTheme, THEMES, type ThemeName } from "../lib/theme";
import Select from "./Select";

export default function ThemeSwitcher(props: { class?: string }) {
    return (
        <Select
            onInput={(e) =>
                setTheme((e.target as HTMLSelectElement).value as ThemeName)
            }
            value={document.documentElement.getAttribute("data-theme") || ""}
            class={props.class}
        >
            {THEMES.map((t) => (
                <option value={t}>{t}</option>
            ))}
        </Select>
    );
}
