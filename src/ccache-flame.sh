#!/bin/bash

source /etc/profile 

function draw_cpu_graph() {
    go tool pprof -seconds=30 -raw -output=ccache-cpu.perf http://127.0.0.1:12121/debug/pprof/profile
    stackcollapse-go.pl ccache-cpu.perf > ccache-cpu.fold
    flamegraph.pl --title="ccache cpu online graph" --colors hot ccache-cpu.fold > ccache-cpu.svg
}

function draw_mem_graph() {
    go tool pprof -alloc_space -raw -output=ccache-mem.perf http://127.0.0.1:12121/debug/pprof/heap
    stackcollapse-go.pl ccache-mem.perf > ccache-mem.fold
    flamegraph.pl --title="ccache mem online graph" --colors mem ccache-mem.fold > ccache-mem.svg
}

draw_cpu_graph
draw_mem_graph
