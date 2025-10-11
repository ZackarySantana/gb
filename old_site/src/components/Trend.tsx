import { onCleanup, onMount, createEffect } from "solid-js";
import uPlot from "uplot";
import "uplot/dist/uPlot.min.css";

export default function Trend(props: {
    series: { x: number[]; y: number[] };
    title: string;
    height?: number;
}) {
    let el!: HTMLDivElement;
    let plot: uPlot | undefined;
    let ro: ResizeObserver | undefined;

    const createPlot = () => {
        const { x, y } = props.series;

        plot = new uPlot(
            {
                title: props.title,
                width: el.clientWidth,
                height: props.height ?? 320,
                pxAlign: 0,
                select: {
                    show: false,
                    left: 0,
                    top: 0,
                    width: 0,
                    height: 0
                },
                scales: { 
                    x: { time: true },
                    y: { auto: true }
                },
                axes: [
                    {
                        grid: { 
                            stroke: "rgba(255,255,255,0.08)",
                            width: 1
                        },
                        ticks: { 
                            stroke: "rgba(255,255,255,0.2)",
                            width: 1
                        },
                        font: "12px system-ui, -apple-system, sans-serif",
                        stroke: "rgba(255,255,255,0.3)"
                    },
                    {
                        grid: { 
                            stroke: "rgba(255,255,255,0.08)",
                            width: 1
                        },
                        ticks: { 
                            stroke: "rgba(255,255,255,0.2)",
                            width: 1
                        },
                        font: "12px system-ui, -apple-system, sans-serif",
                        stroke: "rgba(255,255,255,0.3)"
                    },
                ],
                series: [
                    {},
                    {
                        label: "value",
                        width: 2,
                        stroke: "#3b82f6",
                        points: { 
                            show: true, 
                            size: 6, 
                            width: 2,
                            stroke: "#3b82f6",
                            fill: "#1e40af"
                        },
                    },
                ],
                cursor: {
                    drag: { x: false, y: false },
                    points: {
                        show: true,
                        size: 8,
                        width: 2,
                        stroke: "#ffffff",
                        fill: "#3b82f6"
                    },
                    show: true,
                    lock: false,
                    focus: {
                        prox: 16
                    }
                },
                legend: {
                    show: false
                },
                hooks: {
                    setCursor: [
                        (u) => {
                            const { left, top, idx } = u.cursor;
                            if (idx != null) {
                                const xVal = u.data[0][idx];
                                const yVal = u.data[1][idx];
                                
                                // Update cursor info display
                                const cursorInfo = document.getElementById('cursor-info');
                                if (cursorInfo && yVal != null) {
                                    // Format the date properly
                                    const date = new Date(xVal * 1000);
                                    const dateStr = date.toLocaleDateString('en-US', { 
                                        month: 'short', 
                                        day: 'numeric', 
                                        year: 'numeric' 
                                    });
                                    const timeStr = date.toLocaleTimeString('en-US', { 
                                        hour: 'numeric', 
                                        minute: '2-digit',
                                        hour12: true 
                                    });
                                    cursorInfo.innerHTML = `
                                        <div>Time: ${dateStr}, ${timeStr}</div>
                                        <div>Value: ${yVal.toFixed(2)}</div>
                                    `;
                                }
                            }
                        }
                    ]
                }
            },
            [x, y],
            el
        );

        ro = new ResizeObserver(() => {
            plot?.setSize({
                width: el.clientWidth,
                height: props.height ?? 320,
            });
        });
        ro.observe(el);
    };

    onMount(() => {
        createPlot();
    });

    // Update plot when props change
    createEffect(() => {
        if (plot) {
            const { x, y } = props.series;
            plot.setData([x, y]);
        }
    });
    
    // Update title when it changes (without recreating the plot)
    createEffect(() => {
        if (plot) {
            // Find the title element and update it directly
            const titleEl = plot.root.querySelector('.u-title');
            if (titleEl) {
                titleEl.textContent = props.title;
            }
        }
    });

    onCleanup(() => {
        ro?.disconnect();
        plot?.destroy();
    });

    return (
        <div class="card">
            <div ref={el} />
            <div id="cursor-info" class="cursor-info" />
        </div>
    );
}