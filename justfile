# Compile the native desktop application.
[group('Development')]
build:
    go mod tidy
    mkdir -p dist
    go build -tags desktop -o dist/just-wizard ./cmd/just-wizard

# Run core tests without graphics dependencies.
[group('Development')]
test:
    go test ./cmd/just-wizard

# Compile and test with native graphics dependencies.
[group('Development')]
test-desktop:
    go test -tags desktop ./cmd/just-wizard

# Launch the desktop Wizard for a project.
[group('Application')]
wizard root=".":
    go run -tags desktop ./cmd/just-wizard --root "{{root}}"

# Launch the original terminal interface.
[group('Application')]
terminal root=".":
    go run ./cmd/just-wizard --root "{{root}}"

# Install the native executable into /usr/local/bin.
[group('Application')]
install: build
    sudo mkdir -p /usr/local/bin
    sudo install -m 0755 dist/just-wizard /usr/local/bin/just-wizard
