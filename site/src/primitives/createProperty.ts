import { Accessor, createSignal } from "solid-js";
import { THEME_CHANGE_EVENT } from "../lib/theme";

export function createProperty(property: string): Accessor<string> {
    const computedStyle = getComputedStyle(document.documentElement);
    const propertyValue = computedStyle.getPropertyValue(property).trim();

    const [value, setValue] = createSignal(propertyValue);

    document.addEventListener(THEME_CHANGE_EVENT, () => {
        const computedStyle = getComputedStyle(document.documentElement);
        setValue(computedStyle.getPropertyValue(property).trim());
    });

    return value;
}
