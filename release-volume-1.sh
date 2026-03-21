#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PUZZLETEA_BIN="${PUZZLETEA_BIN:-$SCRIPT_DIR/puzzletea}"
OUT_DIR="$SCRIPT_DIR/out"
PARTS_DIR="$OUT_DIR/parts"
MERGED_JSONL="$OUT_DIR/puzzletea-volume-1.jsonl"
PDF_OUTPUT="$OUT_DIR/puzzletea-volume-1.0.pdf"
BASE_SEED="${BASE_SEED:-tff 2026}"
PDF_TITLE="${PDF_TITLE:-TFF 2026 Preview}"
PDF_ADVERT="${PDF_ADVERT:-Generated via a custom-coded puzzle engine, this collection is a modern tribute to the legacy of Nikoli. Created by Dami Etoile.}"
PDF_SHEET_LAYOUT="${PDF_SHEET_LAYOUT:-duplex-booklet}"

EXPECTED_CATEGORY_COUNT=13
MAX_PUZZLE_PAGES=8
EXPECTED_TOTAL_COUNT=$(((MAX_PUZZLE_PAGES * 4) + 2))

if [[ ! -x "$PUZZLETEA_BIN" ]]; then
  echo "expected built executable at $PUZZLETEA_BIN" >&2
  echo "build it first, for example: go build -o \"$SCRIPT_DIR/puzzletea\"" >&2
  exit 1
fi

mkdir -p "$OUT_DIR" "$PARTS_DIR"

slugify() {
  printf '%s' "$1" | tr '[:upper:]' '[:lower:]' | sed \
    -e 's/+/-plus-/g' \
    -e 's/[^a-z0-9]/-/g' \
    -e 's/--*/-/g' \
    -e 's/^-//' \
    -e 's/-$//'
}

set_modes_for_game() {
  case "$1" in
    "Fillomino")
      MODES=("Mini 5x5" "Easy 6x6" "Medium 8x8" "Hard 10x10" "Expert 12x12")
      ;;
    "Hashiwokakero")
      MODES=(
        "Easy 7x7" "Medium 7x7" "Hard 7x7"
        "Easy 9x9" "Medium 9x9" "Hard 9x9"
        "Easy 11x11" "Medium 11x11" "Hard 11x11"
        "Easy 13x13" "Medium 13x13" "Hard 13x13"
      )
      ;;
    "Hitori")
      MODES=("Mini" "Easy" "Medium" "Tricky" "Hard" "Expert")
      ;;
    "Nonogram")
      MODES=("Mini" "Pocket" "Teaser" "Standard" "Classic" "Tricky" "Large" "Grand" "Epic" "Massive")
      ;;
    "Nurikabe")
      MODES=("Mini" "Easy" "Medium" "Hard" "Expert")
      ;;
    "Ripple Effect")
      MODES=("Mini 5x5" "Easy 6x6" "Medium 7x7" "Hard 8x8" "Expert 9x9")
      ;;
    "Shikaku")
      MODES=("Mini 5x5" "Easy 7x7" "Medium 8x8" "Hard 10x10" "Expert 12x12")
      ;;
    "Sudoku")
      MODES=("Beginner" "Easy" "Medium" "Hard" "Expert" "Diabolical")
      ;;
    "Sudoku RGB")
      MODES=("Beginner" "Easy" "Medium" "Hard" "Expert" "Diabolical")
      ;;
    "Takuzu")
      MODES=("Beginner" "Easy" "Medium" "Tricky" "Hard" "Very Hard" "Extreme")
      ;;
    "Takuzu+")
      MODES=("Beginner" "Easy" "Medium" "Tricky" "Hard" "Very Hard" "Extreme")
      ;;
    "Word Search")
      MODES=("Easy 10x10" "Medium 15x15" "Hard 20x20")
      ;;
    *)
      echo "unknown game manifest entry: $1" >&2
      exit 1
      ;;
  esac
}

allocate_counts() {
  local total="$1"
  local mode_count="$2"
  local i
  local base
  local remainder
  local count

  ALLOC_COUNTS=()
  if (( mode_count <= 0 )); then
    return
  fi

  base=$((total / mode_count))
  remainder=$((total % mode_count))
  for ((i = 0; i < mode_count; i++)); do
    count="$base"
    if (( i < remainder )); then
      count=$((count + 1))
    fi
    ALLOC_COUNTS+=("$count")
  done
}

allocate_category_targets() {
  local total="$1"
  local category_count="$2"

  CATEGORY_TARGETS=()
  if (( category_count <= 0 )); then
    return
  fi

  allocate_counts "$total" "$category_count"
  CATEGORY_TARGETS=("${ALLOC_COUNTS[@]}")
}

bucket_targets_for_total() {
  local total="$1"

  case "$total" in
    6)
      BUCKET_TARGETS=(3 2 1)
      ;;
    5)
      BUCKET_TARGETS=(3 1 1)
      ;;
    4)
      BUCKET_TARGETS=(2 1 1)
      ;;
    3)
      BUCKET_TARGETS=(1 1 1)
      ;;
    2)
      BUCKET_TARGETS=(0 1 1)
      ;;
    *)
      echo "unsupported per-category target ${total}" >&2
      exit 1
      ;;
  esac
}

append_bucket_exports() {
  local game="$1"
  local game_slug="$2"
  local bucket="$3"
  local target="$4"
  local category_file="$5"
  shift 5

  local bucket_modes=("$@")
  local mode_count="${#bucket_modes[@]}"
  local i
  local count
  local mode
  local mode_slug
  local seed
  local temp_file

  if (( mode_count == 0 || target == 0 )); then
    return
  fi

  allocate_counts "$target" "$mode_count"
  for ((i = 0; i < mode_count; i++)); do
    count="${ALLOC_COUNTS[$i]}"
    if (( count == 0 )); then
      continue
    fi

    mode="${bucket_modes[$i]}"
    mode_slug="$(slugify "$mode")"
    seed="${BASE_SEED}:${game_slug}:${bucket}:${mode_slug}"
    temp_file="$(mktemp "${TMPDIR:-/tmp}/${game_slug}-${bucket}-${mode_slug}.XXXXXX.jsonl")"

    echo "Generating ${count} ${bucket} puzzle(s) for ${game} / ${mode}"
    "$PUZZLETEA_BIN" new "$game" "$mode" -e "$count" -o "$temp_file" --with-seed "$seed"
    cat "$temp_file" >> "$category_file"
    rm -f "$temp_file"
  done
}

generate_category_pack() {
  local game="$1"
  local target_total="$2"
  local game_slug
  local category_file
  local total_modes
  local i
  local bucket_index

  local easy_modes=()
  local medium_modes=()
  local hard_modes=()

  set_modes_for_game "$game"
  game_slug="$(slugify "$game")"
  category_file="$PARTS_DIR/${game_slug}.jsonl"
  : > "$category_file"

  total_modes="${#MODES[@]}"
  for ((i = 0; i < total_modes; i++)); do
    bucket_index=$((i * 3 / total_modes))
    case "$bucket_index" in
      0)
        easy_modes+=("${MODES[$i]}")
        ;;
      1)
        medium_modes+=("${MODES[$i]}")
        ;;
      2)
        hard_modes+=("${MODES[$i]}")
        ;;
      *)
        echo "invalid bucket index ${bucket_index} for ${game}" >&2
        exit 1
        ;;
    esac
  done

  bucket_targets_for_total "$target_total"

  append_bucket_exports "$game" "$game_slug" "easy" "${BUCKET_TARGETS[0]}" "$category_file" "${easy_modes[@]}"
  append_bucket_exports "$game" "$game_slug" "medium" "${BUCKET_TARGETS[1]}" "$category_file" "${medium_modes[@]}"
  append_bucket_exports "$game" "$game_slug" "hard" "${BUCKET_TARGETS[2]}" "$category_file" "${hard_modes[@]}"

  local line_count
  line_count="$(wc -l < "$category_file" | tr -d '[:space:]')"
  if [[ "$line_count" != "$target_total" ]]; then
    echo "expected ${target_total} puzzles for ${game}, got ${line_count}" >&2
    exit 1
  fi
}

render_pdf() {
  local cmd=(
    "$PUZZLETEA_BIN"
    export-pdf
    -o "$PDF_OUTPUT"
    --volume 1
    --title "$PDF_TITLE"
    --shuffle-seed "$BASE_SEED"
    --sheet-layout "$PDF_SHEET_LAYOUT"
  )
  local game

  for game in "${GAMES[@]}"; do
    cmd+=("$PARTS_DIR/$(slugify "$game").jsonl")
  done

  if [[ -n "${PDF_HEADER:-}" ]]; then
    cmd+=(--header "$PDF_HEADER")
  fi
  if [[ -n "${PDF_ADVERT:-}" ]]; then
    cmd+=(--advert "$PDF_ADVERT")
  fi

  "${cmd[@]}"
}

GAMES=(
  "Fillomino"
  "Hashiwokakero"
  "Hitori"
  "Nonogram"
  "Nurikabe"
  "Ripple Effect"
  "Shikaku"
  "Sudoku"
  "Sudoku RGB"
  "Takuzu"
  "Takuzu+"
  "Word Search"
)

: > "$MERGED_JSONL"

allocate_category_targets "$EXPECTED_TOTAL_COUNT" "${#GAMES[@]}"

for i in "${!GAMES[@]}"; do
  generate_category_pack "${GAMES[$i]}" "${CATEGORY_TARGETS[$i]}"
done

for game in "${GAMES[@]}"; do
  cat "$PARTS_DIR/$(slugify "$game").jsonl" >> "$MERGED_JSONL"
done

line_count="$(wc -l < "$MERGED_JSONL" | tr -d '[:space:]')"
if [[ "$line_count" != "$EXPECTED_TOTAL_COUNT" ]]; then
  echo "expected ${EXPECTED_TOTAL_COUNT} puzzles in merged jsonl, got ${line_count}" >&2
  exit 1
fi

render_pdf

echo "Wrote $MERGED_JSONL"
echo "Wrote $PDF_OUTPUT"
