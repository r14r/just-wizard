# Architecture

- discovery.go: invokes Just's JSON dump with a fixed justfile and working directory; decodes recipes, groups, modules and positional/option metadata.
- desktop.go: Fyne RUN accordion, EXPLORER tab, command details and output console.
- main.go / terminal.go: numbered terminal fallback.
- explorer*.go: migrated editor, preview, file tree and atomic saving.
- No dependency on the Advent book catalog.
- No recipe execution during discovery. Run only recipes in trusted projects. Confirmation heuristics are not a sandbox.
- JSON schema support is limited to recognized fields; unsupported Just features are delegated to the terminal. No dependency lockfile has been generated in this migration environment. Run go mod tidy and commit go.sum after successful local verification.
