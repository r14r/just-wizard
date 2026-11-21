package main

import (
 "strings"
 "testing"
)

func TestVersionString(t *testing.T) {
 old:=buildVersion
 defer func(){buildVersion=old}()
 buildVersion="0.1.2"
 if got:=versionString();got!="just-wizard 0.1.2"{t.Fatalf("unexpected version output: %q",got)}
 buildVersion="dev"
 if got:=versionString();!strings.HasSuffix(got," dev"){t.Fatalf("unexpected developer version: %q",got)}
}
