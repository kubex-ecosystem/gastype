#!/usr/bin/env bash
# shellcheck disable=SC2038

set -e

PROJECT_PATH="$1"
OUTPUT_DIR="./benchmark_results"
BIN_ORIGINAL="$OUTPUT_DIR/original_bin"
BIN_TRANSPILED="$OUTPUT_DIR/transpiled_bin"

mkdir -p "$OUTPUT_DIR"

echo "🚀 Benchmarking GASType..."
echo "📁 Projeto: $PROJECT_PATH"
echo "📂 Resultados: $OUTPUT_DIR"
echo

# 1️⃣ Build original
echo "🛠  Build original..."
START=$(date +%s%N)
go build -o "$BIN_ORIGINAL" "$PROJECT_PATH"
END=$(date +%s%N)
BUILD_TIME_ORIG=$(( (END - START) / 1000000 ))

SIZE_ORIG=$(stat -c %s "$BIN_ORIGINAL")

# 2️⃣ Transpile + build transpilado
echo "🔄 Transpiling project..."
./gastype transpile -i "$PROJECT_PATH" -o "$OUTPUT_DIR/transpiled" --map "$OUTPUT_DIR/map.json"

echo "🛠  Build transpiled..."
START=$(date +%s%N)
go build -o "$BIN_TRANSPILED" "$OUTPUT_DIR/transpiled"
END=$(date +%s%N)
BUILD_TIME_TRANS=$(( (END - START) / 1000000 ))

SIZE_TRANS=$(stat -c %s "$BIN_TRANSPILED")

# 3️⃣ Startup time
measure_startup() {
    local bin="$1"
    START=$(date +%s%N)
    "$bin" --help >/dev/null 2>&1 || true
    END=$(date +%s%N)
    echo $(( (END - START) / 1000000 ))
}

STARTUP_ORIG=$(measure_startup "$BIN_ORIGINAL")
STARTUP_TRANS=$(measure_startup "$BIN_TRANSPILED")

# 4️⃣ Memória inicial
measure_mem() {
    local bin="$1"
    /usr/bin/time -f "%M" "$bin" --help >/dev/null 2>&1 || true
}

MEM_ORIG=$(measure_mem "$BIN_ORIGINAL")
MEM_TRANS=$(measure_mem "$BIN_TRANSPILED")

# 5️⃣ Linhas de código transpiladas
LINES_TRANS=$(find "$OUTPUT_DIR/transpiled" -name '*.go' | xargs cat | wc -l)
LINES_ORIG=$(find "$PROJECT_PATH" -name '*.go' | xargs cat | wc -l)

# 6️⃣ Resultado
echo "📊 Benchmark Results"
echo "──────────────────────────────────────"
printf "Binary Size:        %8d KB → %d KB\n" $((SIZE_ORIG/1024)) $((SIZE_TRANS/1024))
printf "Build Time:         %8d ms → %d ms\n" $BUILD_TIME_ORIG $BUILD_TIME_TRANS
printf "Startup Time:       %8d ms → %d ms\n" "$STARTUP_ORIG" "$STARTUP_TRANS"
printf "Memory Usage:       %8d KB → %d KB\n" "$MEM_ORIG" "$MEM_TRANS"
printf "Lines of Code:      %8d → %d\n" "$LINES_ORIG" "$LINES_TRANS"
echo "──────────────────────────────────────"
echo "✅ Benchmark completed successfully!"