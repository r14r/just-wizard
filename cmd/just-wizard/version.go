package main

// buildVersion is set from the repository's VERSION file via -ldflags when
// building both desktop and terminal binaries through the Justfile.
var buildVersion = "dev"
