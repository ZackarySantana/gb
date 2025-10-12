import { JSX } from "solid-js";

export default function Select(
    props: JSX.SelectHTMLAttributes<HTMLSelectElement>
) {
    const { class: className, ...rest } = props;

    return (
        <select
            class={`px-4 py-2 rounded border border-border bg-bg-elevated text-text-secondary text-fg outline-none ${className}`}
            {...rest}
        />
    );
}
