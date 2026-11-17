[group('Development')]
# Compile the native desktop application.
build:
    go mod tidy
    mkdir -p dist
    go build -tags desktop -o dist/just-wizard ./cmd/just-wizard

[group('Development')]
# Run core tests without graphics dependencies.
test:
    go test ./cmd/just-wizard

[group('Development')]
# Compile and test with native graphics dependencies.
test-desktop:
    go test -tags desktop ./cmd/just-wizard

[group('Application')]
# Launch the desktop Wizard for a project.
wizard root=".":
    go run -tags desktop ./cmd/just-wizard --root "{{root}}"

[group('Application')]
# Launch the original terminal interface.
terminal root=".":
    go run ./cmd/just-wizard --root "{{root}}"

[group('Application')]
# Install the native executable into GOBIN or GOPATH/bin.
install:
    go install -tags desktop ./cmd/just-wizard
