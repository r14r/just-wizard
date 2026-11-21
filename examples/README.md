# Just Wizard examples

Open the complete example with:

```bash
just-wizard --root examples
```

Or run the individual Just recipes directly from the repository root:

```bash
just --justfile examples/justfile options-books-by-nr
just --justfile examples/justfile options-engines
just --justfile examples/justfile options-layouts
just --justfile examples/justfile build-book 26 modern
just --justfile examples/justfile regenerate-images-book 26 --engine=local
just --justfile examples/justfile test unit
just --justfile examples/justfile deploy staging
```

## Dynamic options recipes

The `options-*` recipes emit **one value per line**. For example:

```just
options-books-by-nr:
    @printf '%s\\n' 1 2 26 42
```

The Wizard runs that recipe when it opens a parameter form:

```just
# @ui param book type=select source=options-books-by-nr default=26
build-book book layout="advent":
    @echo "Building book {{book}} with layout {{layout}}"
```

`source=options-books-by-nr` selects from the output of the command. `options=all|unit|integration`, in contrast, defines **static** choices directly in the annotation.

The `regenerate-images-book` example demonstrates **two dynamic selectors** and mapping an additional UI-only selection into a command-line argument:

```just
# @ui param book type=select source=options-books-by-nr default=26
# @ui param selected_engine type=select source=options-engines flag=--engine default=local
[positional-arguments]
regenerate-images-book book *options:
    @echo "$1 ${@:2}"
```

Selecting book `26` and engine `local` executes:

```bash
just regenerate-images-book 26 --engine=local
```

A UI-only selector requires `flag=--NAME` and a star-variadic recipe parameter (`*options`) to accept the additional argument. The example only prints the resulting arguments; it does not generate images or deploy anything.

**Note:** `options-*` recipes are executable Just commands, not mere configuration. Only use trusted example/project files.
