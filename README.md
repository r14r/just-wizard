# Just Wizard

Standalone Go/Fyne desktop frontend for Just. Extracted from the Advent application without its book data, catalog, or project-specific pipeline.

## Start

Install Go 1.23+, Just with JSON dump support, and Fyne's native CGO toolchain. On macOS run `xcode-select --install`. On Debian/Ubuntu install gcc, libgl1-mesa-dev and xorg-dev.

```sh
just
just doctor
just ci
just test-desktop
just build
./dist/just-wizard --root /path/to/project
```

After `just install`, run `just-wizard` from any project directory (installed at /usr/local/bin/just-wizard). Use `--justfile path` for a custom filename. `just terminal` launches the non-graphical variant.

## Project commands

Run `just` to list all recipes. `just version` prints VERSION; `just bootstrap` resolves Go modules and updates `go.sum` using `go mod tidy`; `just ci` validates the justfile and runs Go vet and tests. `just ci-desktop` also tests and builds the native GUI. `just build-terminal` creates the terminal executable, `just install-terminal` installs it alongside the desktop binary, and `just clean` removes local build output.

## Dynamic navigation

At startup the application finds justfile, Justfile or .justfile in the selected directory and reads the compiled JSON dump using Just itself. It does not execute recipes to build the menu.

```just
# Run all tests.
[group('Development')]
test:
    go test ./...

# Build one edition.
[group('Publishing')]
build-book book layout="advent":
    echo "{{book}} {{layout}}"
```

RUN shows actual groups and recipes. Ungrouped recipes appear under Ungrouped. Nested module recipes use qualified names; local aliases are supported. Parameter forms and descriptions derive from the dump. Defaults are applied by Just if optional trailing arguments are omitted. Variadic fields split on whitespace. Complex list/argument features may require the terminal. Restart after editing the justfile.

EXPLORER replaces the sidebar with a project file tree. It retains text editing, JSON formatting, YAML/JSON/XML validation, Markdown source/live preview, image previews, safe saving, conflict detection and right-click Open/View. PDF previews require Poppler; HTML previews require Chromium or Chrome. HTML rendering executes page scripts, so preview only trusted files. The blue output console shows completed command output, with Copy/Clear/Cancel. Interactive recipes should use the terminal version. Cancellation targets the Just process, not necessarily every descendant.

The original Advent repository remains unchanged. Its specific book-management and Getting Started workflows are not part of this general-purpose app; define project-specific workflows as Just recipes.

See [technical notes](docs/ARCHITECTURE.md). Native compilation and UI verification are required.

## Rich UI metadata and migration

Run `just-wizard migrate --dry-run` to preview annotations, then `just-wizard migrate` to add missing annotations to the current Justfile. Use `--root` or `--justfile` for another project. Migration validates the proposed file with Just, preserves recipe bodies, and writes a `.bak` backup before replacing the file. Move an existing backup before another write. Existing annotations are preserved; automatic type/group inference is conservative and never invents dropdown options or numeric bounds. Only simple root-level recipe headers are annotated; modules and imported recipes remain discoverable through Just's JSON dump.

Supported comments match just-ui: `@ui group`, `@ui label`, `@ui description`, `@ui confirm`, and `@ui param NAME type=select options=a|b default=a label=Display_name description=Help_text`. Also supported: text, password, number (`min`/`max`), path (text input), checkbox, and `required`. Explicit UI groups override native groups. Just remains authoritative for executable signatures. Password previews hide arguments; recipe output may still expose secrets, and terminal input is not masked. Restart the UI after changing metadata.

### Recipe icons

Use `# @ui icon hammer` to override the icon of a recipe. Supported IDs: `terminal`, `package`, `hammer`, `flask-conical`, `clipboard-check`, `activity`, `trash-2`, `download`, `upload`, `tag`, `monitor`, `list`, `settings`, `square`, `refresh-cw`, `archive`, `play`. Migration adds a missing icon using the recognized command name/prefix first, then the effective group, or `play` as fallback. Existing icon annotations are preserved; unsupported values fall back to automatic recognition. Both binaries share the inference logic. No changes to recipe signatures or execution are required.

## CLI options reference

Use `just-wizard options` to list supported `@ui` annotations, parameter types and attributes, all icon IDs, and a copyable Justfile example. `just-wizard options --example` prints just the example; `just-wizard options --icons` lists icon IDs. These commands do not require a Justfile or start the UI, and work in both desktop and terminal builds.

### Go dependency checksums

The repository requires a `go.sum` file for reproducible desktop builds. If it is missing, run `just bootstrap` (or `go mod tidy`) before building, then commit the generated `go.sum` alongside `go.mod`. `just build` runs `bootstrap` automatically.

## Executable version

`just-wizard version` and `just-wizard --version` are aliases and print the version embedded during `just build` or `just build-terminal`, for example `just-wizard 0.1.2`. Both work without a project Justfile or desktop UI. Developer binaries built without the project's Just recipes report `dev` unless the same version linker flag is provided.


### Dynamic selector values

In addition to static `options=a|b|c`, a select parameter can use
`source=options-<name>` to populate values by running a **project-defined**
Just recipe when its form opens:

```just
options-books-by-nr:
    @printf '%s\\n' 1 2 3

# @ui param book type=select label=Book source=options-books-by-nr
build-book book:
    @echo "{{book}}"
```

Dynamic source recipes must be named `options-*`, present in the discovered
Justfile and return one value per line. The wizard executes the source
without a shell, using the selected `--justfile` and project working
directory, with an eight-second timeout, bounded output and deduplication.
Invalid/empty output and execution failures are displayed as errors.
Only use trusted project Justfiles: a source is an executable recipe.

For `*options` recipes, parameters not defined by Just may be declared
as UI-only selectors with an explicit `flag=--...` mapping:

```just
options-engines:
    @printf '%s\\n' local stable-diffusion huggingface

# @ui param book type=select source=options-books-by-nr
# @ui param selected_engine type=select label=Engine source=options-engines flag=--engine
[positional-arguments]
regenerate-images-book book *options:
    @echo "$1 ${@:2}"
```

Selecting book `26` and engine `local` executes
`just regenerate-images-book 26 --engine=local`. Static `options=` and
existing input types remain supported. UI-only parameters require an explicit
flag and a star-variadic Just parameter. Use `just-wizard options` for details.

## Runnable examples

See [examples/justfile](examples/justfile) and [examples/README.md](examples/README.md) for static `options=` lists, dynamic `source=options-*` lists, and forwarding UI-only selections as CLI flags. Open with `just-wizard --root examples` or run `just --justfile examples/justfile options-books-by-nr` to inspect provider values.
