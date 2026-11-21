# Just Wizard – standalone Go/Fyne desktop and terminal application.
set shell := ["bash", "-cu"]

# @ui group Information
# @ui icon list
# @ui label Default
# @ui description Run the default recipe.
default:
    @just --list

# @ui group Information
# @ui icon tag
# @ui label Version
# @ui description Print the project version.
# Print the project version.
version:
    @cat VERSION

# @ui group Setup
# @ui icon package
# @ui label Bootstrap
# @ui description Resolve Go module dependencies and create missing go.sum checksums.
# Resolve Go module dependencies and create missing go.sum checksums.
bootstrap:
    go mod tidy

# @ui group Development
# @ui icon activity
# @ui label Doctor
# @ui description Check the Go and Just toolchain.
# Check the Go and Just toolchain.
doctor:
    go version
    just --version
    just --dump --dump-format json > /dev/null
    go env GOOS GOARCH CGO_ENABLED CC

# @ui group Development
# @ui icon clipboard-check
# @ui label Ci
# @ui description Validate the justfile and run Go checks without graphics dependencies.
# Validate the justfile and run Go checks without graphics dependencies.
ci: bootstrap
    just --dump --dump-format json > /dev/null
    go vet ./...
    go test ./...

# @ui group Build
# @ui icon hammer
# @ui label Build
# @ui description Compile the native desktop application.
# Compile the native desktop application.
build: bootstrap
    mkdir -p dist
    go build -tags desktop -ldflags "-X main.buildVersion=$(cat VERSION)" -o dist/just-wizard ./cmd/just-wizard

# @ui group Build
# @ui icon hammer
# @ui label Build terminal
# @ui description Compile the terminal-only application.
# Compile the terminal-only application.
build-terminal: bootstrap
    mkdir -p dist
    go build -ldflags "-X main.buildVersion=$(cat VERSION)" -o dist/just-wizard-terminal ./cmd/just-wizard

# @ui group Development
# @ui icon flask-conical
# @ui label Test
# @ui description Run core tests without graphics dependencies.
# Run core tests without graphics dependencies.
test: bootstrap
    go test ./...

# @ui group Development
# @ui icon flask-conical
# @ui label Test desktop
# @ui description Compile and test with native graphics dependencies.
# Compile and test with native graphics dependencies.
test-desktop: bootstrap
    go test -tags desktop ./...

# @ui group Development
# @ui icon clipboard-check
# @ui label Ci desktop
# @ui description Validate the native desktop variant.
# Validate the native desktop variant.
ci-desktop: ci test-desktop build

# @ui group Application
# @ui icon monitor
# @ui label Wizard
# @ui description Launch the desktop Wizard for a project.
# @ui param root type=path label=root
# Launch the desktop Wizard for a project.
wizard root=".": build
    ./dist/just-wizard --root {{quote(root)}}

# @ui group Application
# @ui icon terminal
# @ui label Terminal
# @ui description Launch the original terminal interface.
# @ui param root type=path label=root
# Launch the original terminal interface.
terminal root=".": build-terminal
    ./dist/just-wizard-terminal --root {{quote(root)}}

# @ui group Application
# @ui icon terminal
# @ui label Wizard terminal
# @ui description Compatibility command matching the Adventskalender project.
# @ui param root type=path label=root
# Compatibility command matching the Adventskalender project.
wizard-terminal root=".":
    just terminal {{quote(root)}}

# @ui group Setup
# @ui icon download
# @ui label Install
# @ui description Install the native executable into /usr/local/bin.
# Install the native executable into /usr/local/bin.
install: build
    sudo mkdir -p /usr/local/bin
    sudo install -m 0755 dist/just-wizard /usr/local/bin/just-wizard
    @echo "Installed: /usr/local/bin/just-wizard"

# @ui group Setup
# @ui icon download
# @ui label Install terminal
# @ui description Install the terminal-only executable alongside the desktop application.
# Install the terminal-only executable alongside the desktop application.
install-terminal: build-terminal
    sudo mkdir -p /usr/local/bin
    sudo install -m 0755 dist/just-wizard-terminal /usr/local/bin/just-wizard-terminal
    @echo "Installed: /usr/local/bin/just-wizard-terminal"

# @ui group Maintenance
# @ui icon trash-2
# @ui label Clean
# @ui description Remove generated local binaries only.
# @ui confirm
# Remove generated local binaries only.
clean:
    rm -rf dist

# @ui group Maintenance
# @ui icon refresh-cw
# @ui label Migrate
# @ui description Add missing UI metadata to the selected project Justfile.
# @ui confirm
# @ui param root type=path label=root
# Add missing UI metadata to the selected project Justfile.
migrate root=".": build-terminal
    ./dist/just-wizard-terminal migrate --root {{quote(root)}}
