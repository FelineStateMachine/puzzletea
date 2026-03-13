# AGENTS

- Validate every change with `just fmt` and `just lint`.
- Prefer `just test-short` for test runs.
- Use the -short flag when testing non-generator code.
- Use full-length tests only when validating long-running generator behavior.
- Package map:
- `catalog` is the pure metadata index.
- `registry` is the concrete built-in runtime registry.
- `gameentry` builds validated runtime entries from definitions plus modes.
- `pdfexport` owns the printable export pipeline.
- `builtinprint` only bootstraps built-in print adapters.
