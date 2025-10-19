/* /{benchmark}/commit/{commit} */

import { useParams } from "@solidjs/router";
import { Component } from "solid-js";

const Benchmarks: Component = () => {
    const params = useParams();

    return (
        <div class="py-8">
            <p>Testing</p>
            <p>Commit: {params.commit}</p>
            <p>Benchmark: {params.benchmark}</p>
        </div>
    );
};

export default Benchmarks;
