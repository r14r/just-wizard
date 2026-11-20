# Just Wizard – standalone Go/Fyne desktop and terminal application.
set shell := ["bash", "-cu"]

default:
    @just --list

# Print the project version.
version:
    @cat VERSION

# Resolve Go module dependencies and create missing go.sum checksums.
bootstrap:
    go mod tidy

# Check the Go and Just toolchain.
doctor:
    go version
    just --version
    just --dump --dump-format json > /dev/null
    go env GOOS GOARCH CGO_ENABLED CC

# Validate the justfile and run Go checks without graphics dependencies.
ci: bootstrap
    just --dump --dump-format json > /dev/null
    go vet ./...
    go test ./...

# Compile the native desktop application.
build: bootstrap
    mkdir -p dist
    go build -tags desktop -o dist/just-wizard ./cmd/just-wizard

# Compile the terminal-only application.
build-terminal: bootstrap
    mkdir -p dist
    go build -o dist/just-wizard-terminal ./cmd/just-wizard

# Run core tests without graphics dependencies.
test: bootstrap
    go test ./...

# Compile and test with native graphics dependencies.
test-desktop: bootstrap
    go test -tags desktop ./...

# Validate the native desktop variant.
ci-desktop: ci test-desktop build

# Launch the desktop Wizard for a project.
wizard root=".": build
    ./dist/just-wizard --root {{quote(root)}}

# Launch the original terminal interface.
terminal root=".": build-terminal
    ./dist/just-wizard-terminal --root {{quote(root)}}

# Compatibility command matching the Adventskalender project.
wizard-terminal root=".":
    just terminal {{quote(root)}}

# Install the native executable into /usr/local/bin.
install: build
    sudo mkdir -p /usr/local/bin
    sudo install -m 0755 dist/just-wizard /usr/local/bin/just-wizard
    @echo "Installed: /usr/local/bin/just-wizard"

# Install the terminal-only executable alongside the desktop application.
install-terminal: build-terminal
    sudo mkdir -p /usr/local/bin
    sudo install -m 0755 dist/just-wizard-terminal /usr/local/bin/just-wizard-terminal
    @echo "Installed: /usr/local/bin/just-wizard-terminal"

# Remove generated local binaries only.
clean:
    rm -rf dist

# Add missing UI metadata to the selected project Justfile.
migrate root=".": build-terminal
    ./dist/just-wizard-terminal migrate --root {{quote(root)}}
