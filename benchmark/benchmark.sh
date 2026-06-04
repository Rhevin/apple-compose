#!/usr/bin/env bash
# benchmark/benchmark.sh
# Benchmarks Apple Container (via apple-compose) vs OrbStack using the same
# methodology as:
# https://www.repoflow.io/blog/apple-containers-vs-docker-desktop-vs-orbstack
#
# Tests (each run RUNS times, results averaged):
#   1. Container startup time
#   2. CPU (sysbench single-thread + all-threads)
#   3. Memory throughput (sysbench)
#   4. Disk I/O (fio, bind mount)
#   5. Small file workflow
#
# Usage: ./benchmark.sh [--runs N] [--skip <test>] [--output results.md]

set -euo pipefail

# ── Config ────────────────────────────────────────────────────────────────────
RUNS=${RUNS:-20}
OUTPUT="${OUTPUT:-results.md}"
SKIP_TESTS=""

APPLE_TMPDIR=""
ORBSTACK_TMPDIR=""

# Resolve apple-compose binary relative to this script's directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
APPLE_COMPOSE="${APPLE_COMPOSE:-$REPO_ROOT/apple-compose}"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.yml"

# ── Colours ───────────────────────────────────────────────────────────────────
R='\033[0;31m' G='\033[0;32m' C='\033[0;36m'
BOLD='\033[1m' DIM='\033[2m' NC='\033[0m'

# ── Arg parsing ───────────────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case $1 in
    --runs)   RUNS="$2";   shift 2 ;;
    --skip)   SKIP_TESTS="$SKIP_TESTS $2"; shift 2 ;;
    --output) OUTPUT="$2"; shift 2 ;;
    *) echo "Unknown arg: $1"; exit 1 ;;
  esac
done

should_skip() { [[ " $SKIP_TESTS " == *" $1 "* ]]; }

# ── Helpers ───────────────────────────────────────────────────────────────────
log()  { echo -e "${DIM}  $*${NC}"; }
info() { echo -e "${C}▶ $*${NC}"; }
ok()   { echo -e "${G}✔ $*${NC}"; }
sep()  { printf '%0.s─' {1..72}; echo; }

average() {
  awk '{s=0; for(i=1;i<=NF;i++) s+=$i; printf "%.4f", s/NF}' <<< "$*"
}

parse_time() {
  # convert bash `time` real output (e.g. "0m1.234s") to seconds
  sed 's/m/:/g' | sed 's/s//g' | awk -F: '{printf "%.4f", $1*60+$2}'
}

# ── Runtime wrappers ──────────────────────────────────────────────────────────
# OrbStack: standard docker CLI with context switch
orbstack_run() {
  docker --context orbstack run --rm "$@"
}

# Apple Container: apple-compose run (uses Apple Container runtime under the hood)
apple_run() {
  "$APPLE_COMPOSE" -f "$COMPOSE_FILE" run bench "$@"
}

# ── Runtime detection ─────────────────────────────────────────────────────────
check_runtimes() {
  HAVE_ORBSTACK=false
  HAVE_APPLE=false

  if docker context inspect orbstack &>/dev/null; then
    HAVE_ORBSTACK=true
  fi

  if [[ -x "$APPLE_COMPOSE" ]]; then
    HAVE_APPLE=true
  fi

  if ! $HAVE_ORBSTACK && ! $HAVE_APPLE; then
    echo -e "${R}Error: neither OrbStack nor apple-compose found.${NC}"
    exit 1
  fi

  echo -e "${BOLD}Detected runtimes:${NC}"
  $HAVE_ORBSTACK && echo -e "  ${G}✔ OrbStack${NC}  (docker context: orbstack)"
  $HAVE_APPLE    && echo -e "  ${G}✔ Apple Container${NC}  (apple-compose: $APPLE_COMPOSE)"
  echo
}

pull_images() {
  info "Pulling images…"
  $HAVE_ORBSTACK && docker --context orbstack pull alpine:3.20 &>/dev/null             && log "OrbStack: pulled alpine:3.20"
  $HAVE_APPLE    && "$APPLE_COMPOSE" -f "$COMPOSE_FILE" pull &>/dev/null               && log "Apple:    pulled alpine:3.20"
}

# ── 1. Startup time ───────────────────────────────────────────────────────────
bench_startup() {
  info "Test 1/5: Container startup time (${RUNS} runs each)"
  local orb_times=() apple_times=()

  for ((i=1; i<=RUNS; i++)); do
    if $HAVE_ORBSTACK; then
      t=$( { time orbstack_run alpine:3.20 true; } 2>&1 | grep real | awk '{print $2}' | parse_time )
      orb_times+=("$t")
    fi
    if $HAVE_APPLE; then
      t=$( { time apple_run true; } 2>&1 | grep real | awk '{print $2}' | parse_time )
      apple_times+=("$t")
    fi
    printf "\r  run %d/%d" "$i" "$RUNS"
  done
  echo

  STARTUP_ORB=$(average "${orb_times[*]:-0}")
  STARTUP_APPLE=$(average "${apple_times[*]:-0}")
  ok "Startup: OrbStack=${STARTUP_ORB}s  Apple=${STARTUP_APPLE}s"
}

# ── 2. CPU ────────────────────────────────────────────────────────────────────
SYSBENCH_CPU="apk add --quiet sysbench && sysbench cpu --max-time=60"

bench_cpu() {
  info "Test 2/5: CPU benchmark (${RUNS} runs each)"
  local orb_single=() apple_single=() orb_multi=() apple_multi=()
  local ncpu
  ncpu=$(sysctl -n hw.logicalcpu 2>/dev/null || nproc)

  for ((i=1; i<=RUNS; i++)); do
    if $HAVE_ORBSTACK; then
      v=$(orbstack_run alpine:3.20 sh -c "$SYSBENCH_CPU --threads=1 run" 2>/dev/null | grep "events per second" | awk '{print $NF}')
      orb_single+=("${v:-0}")
      v=$(orbstack_run alpine:3.20 sh -c "$SYSBENCH_CPU --threads=$ncpu run" 2>/dev/null | grep "events per second" | awk '{print $NF}')
      orb_multi+=("${v:-0}")
    fi
    if $HAVE_APPLE; then
      v=$(apple_run sh -c "$SYSBENCH_CPU --threads=1 run" 2>/dev/null | grep "events per second" | awk '{print $NF}')
      apple_single+=("${v:-0}")
      v=$(apple_run sh -c "$SYSBENCH_CPU --threads=$ncpu run" 2>/dev/null | grep "events per second" | awk '{print $NF}')
      apple_multi+=("${v:-0}")
    fi
    printf "\r  run %d/%d" "$i" "$RUNS"
  done
  echo

  CPU_ORB_SINGLE=$(average "${orb_single[*]:-0}")
  CPU_ORB_MULTI=$(average "${orb_multi[*]:-0}")
  CPU_APPLE_SINGLE=$(average "${apple_single[*]:-0}")
  CPU_APPLE_MULTI=$(average "${apple_multi[*]:-0}")
  ok "CPU single: OrbStack=${CPU_ORB_SINGLE} ev/s  Apple=${CPU_APPLE_SINGLE} ev/s"
  ok "CPU multi:  OrbStack=${CPU_ORB_MULTI} ev/s  Apple=${CPU_APPLE_MULTI} ev/s"
}

# ── 3. Memory ─────────────────────────────────────────────────────────────────
SYSBENCH_MEM="apk add --quiet sysbench && sysbench memory --memory-block-size=1M --memory-total-size=4G"

bench_memory() {
  info "Test 3/5: Memory throughput (${RUNS} runs each)"
  local ncpu
  ncpu=$(sysctl -n hw.logicalcpu 2>/dev/null || nproc)
  local orb_vals=() apple_vals=()

  for ((i=1; i<=RUNS; i++)); do
    if $HAVE_ORBSTACK; then
      v=$(orbstack_run alpine:3.20 sh -c "$SYSBENCH_MEM --threads=$ncpu run" 2>/dev/null | grep "transferred" | grep -oE '[0-9]+\.[0-9]+ MiB/sec' | awk '{print $1}')
      orb_vals+=("${v:-0}")
    fi
    if $HAVE_APPLE; then
      v=$(apple_run sh -c "$SYSBENCH_MEM --threads=$ncpu run" 2>/dev/null | grep "transferred" | grep -oE '[0-9]+\.[0-9]+ MiB/sec' | awk '{print $1}')
      apple_vals+=("${v:-0}")
    fi
    printf "\r  run %d/%d" "$i" "$RUNS"
  done
  echo

  MEM_ORB=$(average "${orb_vals[*]:-0}")
  MEM_APPLE=$(average "${apple_vals[*]:-0}")
  ok "Memory: OrbStack=${MEM_ORB} MiB/s  Apple=${MEM_APPLE} MiB/s"
}

# ── 4. Disk I/O ───────────────────────────────────────────────────────────────
FIO_CMD="apk add --quiet fio && \
  fio --name=seq-read --rw=read --bs=1M --size=1G --numjobs=1 \
      --time_based=0 --filename=/mnt/testfile --ioengine=sync \
      --output-format=terse 2>/dev/null"

bench_disk() {
  info "Test 4/5: Disk I/O — fio seq-read on bind mount (${RUNS} runs each)"
  local orb_vals=() apple_vals=()

  ORBSTACK_TMPDIR=$(mktemp -d)
  APPLE_TMPDIR=$(mktemp -d)

  for ((i=1; i<=RUNS; i++)); do
    if $HAVE_ORBSTACK; then
      v=$(orbstack_run -v "${ORBSTACK_TMPDIR}:/mnt" alpine:3.20 sh -c "$FIO_CMD" 2>/dev/null | sed 's/\r/\n/g' | grep "^3;" | head -1 | cut -d';' -f7)
      orb_vals+=("${v:-0}")
    fi
    if $HAVE_APPLE; then
      v=$(BENCH_MOUNT="$APPLE_TMPDIR" apple_run sh -c "$FIO_CMD" 2>/dev/null | sed 's/\x1b\[[0-9;?]*[mhlKHJ]//g' | sed 's/\r/\n/g' | grep "^3;" | head -1 | cut -d';' -f7)
      apple_vals+=("${v:-0}")
    fi
    printf "\r  run %d/%d" "$i" "$RUNS"
  done
  echo

  DISK_ORB=$(average "${orb_vals[*]:-0}")
  DISK_APPLE=$(average "${apple_vals[*]:-0}")
  DISK_ORB_MIB=$(awk "BEGIN {printf \"%.2f\", $DISK_ORB/1024}")
  DISK_APPLE_MIB=$(awk "BEGIN {printf \"%.2f\", $DISK_APPLE/1024}")
  ok "Disk seq-read: OrbStack=${DISK_ORB_MIB} MiB/s  Apple=${DISK_APPLE_MIB} MiB/s"
}

# ── 5. Small file workflow ────────────────────────────────────────────────────
SMALLFILE_CMD='set -e
DIR=/mnt/smallfiles
mkdir -p $DIR
for i in $(seq 1 1000); do echo "data-$i" > $DIR/file$i.txt; done
for i in $(seq 1 1000); do cat $DIR/file$i.txt > /dev/null; done
for i in $(seq 1 1000); do stat $DIR/file$i.txt > /dev/null; done
ls $DIR > /dev/null
cp -r $DIR ${DIR}_copy
rm -rf $DIR ${DIR}_copy'

bench_smallfiles() {
  info "Test 5/5: Small file workflow — 1000 files (${RUNS} runs each)"
  local orb_vals=() apple_vals=()

  [[ -z "$ORBSTACK_TMPDIR" ]] && ORBSTACK_TMPDIR=$(mktemp -d)
  [[ -z "$APPLE_TMPDIR" ]]   && APPLE_TMPDIR=$(mktemp -d)

  for ((i=1; i<=RUNS; i++)); do
    if $HAVE_ORBSTACK; then
      t=$( { time orbstack_run -v "${ORBSTACK_TMPDIR}:/mnt" alpine:3.20 sh -c "$SMALLFILE_CMD"; } 2>&1 | \
           grep real | awk '{print $2}' | parse_time )
      orb_vals+=("${t:-0}")
    fi
    if $HAVE_APPLE; then
      t=$( { time BENCH_MOUNT="$APPLE_TMPDIR" apple_run sh -c "$SMALLFILE_CMD"; } 2>&1 | \
           grep real | awk '{print $2}' | parse_time )
      apple_vals+=("${t:-0}")
    fi
    printf "\r  run %d/%d" "$i" "$RUNS"
  done
  echo

  SMALLFILE_ORB=$(average "${orb_vals[*]:-0}")
  SMALLFILE_APPLE=$(average "${apple_vals[*]:-0}")
  ok "Small files: OrbStack=${SMALLFILE_ORB}s  Apple=${SMALLFILE_APPLE}s"
}

# ── Results ───────────────────────────────────────────────────────────────────
winner() {
  local orb=$1 apple=$2 higher=${3:-true}
  if $higher; then
    awk -v o="$orb" -v a="$apple" 'BEGIN { print (o+0 >= a+0) ? "OrbStack" : "Apple" }'
  else
    awk -v o="$orb" -v a="$apple" 'BEGIN { print (o+0 <= a+0) ? "OrbStack" : "Apple" }'
  fi
}

print_results() {
  echo
  echo -e "${BOLD}$(sep)${NC}"
  echo -e "${BOLD}  RESULTS  (averaged over ${RUNS} runs)${NC}"
  echo -e "${BOLD}$(sep)${NC}"
  printf "  %-36s %14s %14s %10s\n" "TEST" "ORBSTACK" "APPLE" "WINNER"
  sep
  printf "  %-36s %14s %14s %10s\n" "Startup time (s, lower=better)"  "${STARTUP_ORB}s"      "${STARTUP_APPLE}s"      "$(winner "$STARTUP_ORB"    "$STARTUP_APPLE"    false)"
  printf "  %-36s %14s %14s %10s\n" "CPU single-thread (ev/s)"        "$CPU_ORB_SINGLE"      "$CPU_APPLE_SINGLE"      "$(winner "$CPU_ORB_SINGLE" "$CPU_APPLE_SINGLE" true)"
  printf "  %-36s %14s %14s %10s\n" "CPU multi-thread (ev/s)"         "$CPU_ORB_MULTI"       "$CPU_APPLE_MULTI"       "$(winner "$CPU_ORB_MULTI"  "$CPU_APPLE_MULTI"  true)"
  printf "  %-36s %14s %14s %10s\n" "Memory throughput (MiB/s)"       "$MEM_ORB"             "$MEM_APPLE"             "$(winner "$MEM_ORB"        "$MEM_APPLE"        true)"
  printf "  %-36s %14s %14s %10s\n" "Disk seq-read (MiB/s)"           "$DISK_ORB_MIB"        "$DISK_APPLE_MIB"        "$(winner "$DISK_ORB_MIB"   "$DISK_APPLE_MIB"   true)"
  printf "  %-36s %14s %14s %10s\n" "Small files (s, lower=better)"   "${SMALLFILE_ORB}s"    "${SMALLFILE_APPLE}s"    "$(winner "$SMALLFILE_ORB"  "$SMALLFILE_APPLE"  false)"
  sep

  cat > "$OUTPUT" <<MD
# Container Benchmark: OrbStack vs Apple Container

**Runs per test:** ${RUNS}
**Image:** \`alpine:3.20\`
**Date:** $(date +%Y-%m-%d)
**Host:** $(uname -m) — $(sysctl -n machdep.cpu.brand_string 2>/dev/null || echo "unknown CPU")

| Test | OrbStack | Apple Container | Winner |
|------|----------|-----------------|--------|
| Startup time (s, lower=better) | ${STARTUP_ORB} | ${STARTUP_APPLE} | $(winner "$STARTUP_ORB" "$STARTUP_APPLE" false) |
| CPU single-thread (ev/s) | ${CPU_ORB_SINGLE} | ${CPU_APPLE_SINGLE} | $(winner "$CPU_ORB_SINGLE" "$CPU_APPLE_SINGLE" true) |
| CPU multi-thread (ev/s) | ${CPU_ORB_MULTI} | ${CPU_APPLE_MULTI} | $(winner "$CPU_ORB_MULTI" "$CPU_APPLE_MULTI" true) |
| Memory throughput (MiB/s) | ${MEM_ORB} | ${MEM_APPLE} | $(winner "$MEM_ORB" "$MEM_APPLE" true) |
| Disk seq-read (MiB/s) | ${DISK_ORB_MIB} | ${DISK_APPLE_MIB} | $(winner "$DISK_ORB_MIB" "$DISK_APPLE_MIB" true) |
| Small files workflow (s, lower=better) | ${SMALLFILE_ORB} | ${SMALLFILE_APPLE} | $(winner "$SMALLFILE_ORB" "$SMALLFILE_APPLE" false) |
MD
  echo -e "\n  ${G}Results written to ${OUTPUT}${NC}"
}

# ── Cleanup ───────────────────────────────────────────────────────────────────
cleanup() {
  [[ -n "$ORBSTACK_TMPDIR" && -d "$ORBSTACK_TMPDIR" ]] && rm -rf "$ORBSTACK_TMPDIR"
  [[ -n "$APPLE_TMPDIR"    && -d "$APPLE_TMPDIR"    ]] && rm -rf "$APPLE_TMPDIR"
}
trap cleanup EXIT

# ── Main ──────────────────────────────────────────────────────────────────────
main() {
  echo -e "\n${BOLD}${C}  Container Runtime Benchmark${NC}"
  echo -e "  ${DIM}OrbStack vs Apple Container — $RUNS runs per test${NC}\n"

  check_runtimes
  pull_images

  STARTUP_ORB=0;  STARTUP_APPLE=0
  CPU_ORB_SINGLE=0; CPU_ORB_MULTI=0; CPU_APPLE_SINGLE=0; CPU_APPLE_MULTI=0
  MEM_ORB=0;      MEM_APPLE=0
  DISK_ORB=0;     DISK_ORB_MIB=0; DISK_APPLE=0; DISK_APPLE_MIB=0
  SMALLFILE_ORB=0; SMALLFILE_APPLE=0

  should_skip startup    || bench_startup
  should_skip cpu        || bench_cpu
  should_skip memory     || bench_memory
  should_skip disk       || bench_disk
  should_skip smallfiles || bench_smallfiles

  print_results
}

main
