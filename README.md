# Just Wizard

Standalone Go/Fyne desktop frontend for Just. Extracted from the Advent application without its book data, catalog, or project-specific pipeline.

## Start

Install Go 1.23+, Just with JSON dump support, and Fyne's native CGO toolchain. On macOS run `xcode-select --install`. On Debian/Ubuntu install gcc, libgl1-mesa-dev and xorg-dev.

```sh
go mod tidy
just test
just test-desktop
just build
./dist/just-wizard --root /path/to/project
```

After `just install`, run `just-wizard` from any project directory (ensure GOBIN is on PATH). Use `--justfile path` for a custom filename. `just terminal` launches the non-graphical variant.

## Dynamic navigation

At startup the application finds justfile, Justfile or .justfile in the selected directory and reads the compiled JSON dump using Just itself. It does not execute recipes to build the menu.

```just
[group('Development')]
# Run all tests.
test:
    go test ./...

[group('Publishing')]
# Build one edition.
build-book book layout="advent":
    echo "{{book}} {{layout}}"
```

RUN shows actual groups and recipes. Ungrouped recipes appear under Ungrouped. Nested module recipes use qualified names; local aliases are supported. Parameter forms and descriptions derive from the dump. Defaults are applied by Just if optional trailing arguments are omitted. Variadic fields split on whitespace. Complex list/argument features may require the terminal. Restart after editing the justfile.

EXPLORER replaces the sidebar with a project file tree. It retains text editing, JSON formatting, YAML/JSON/XML validation, Markdown source/live preview, image previews, safe saving, conflict detection and right-click Open/View. PDF previews require Poppler; HTML previews require Chromium or Chrome. HTML rendering executes page scripts, so preview only trusted files. The blue output console shows completed command output, with Copy/Clear/Cancel. Interactive recipes should use the terminal version. Cancellation targets the Just process, not necessarily every descendant.

The original Advent repository remains unchanged. Its specific book-management and Getting Started workflows are not part of this general-purpose app; define project-specific workflows as Just recipes.

See [technical notes](docs/ARCHITECTURE.md). Native compilation and UI verification are required.
