package main

import (
 "bytes"
 "strings"
 "testing"
)
func TestOptionsHelp(t *testing.T){
 var b bytes.Buffer
 printOptions(&b,false,false)
 for _,s:=range []string{"@ui group","@ui label","@ui description","@ui icon","@ui confirm","@ui param","type=text|password|number|path|select|checkbox","options=a|b|c","min=NUMBER","max=NUMBER","Example Justfile:"}{
  if !strings.Contains(b.String(),s){t.Errorf("missing %q",s)}
 }
}
func TestOptionsExampleAndIcons(t *testing.T){
 var b bytes.Buffer
 printOptions(&b,true,false)
 if !strings.Contains(b.String(),"test suite=")||strings.Contains(b.String(),"Usage:"){t.Fatal("unexpected example output")}
 b.Reset()
 printOptions(&b,false,true)
 if !strings.Contains(b.String(),"flask-conical")||strings.Contains(b.String(),"Example Justfile:"){t.Fatal("unexpected icon output")}
}
func TestOptionsRejectInvalidArgs(t *testing.T){
 for _,args:=range [][]string{{"--example","--icons"},{"unexpected"}} {
  if err:=runOptions(args);err==nil {t.Errorf("expected error for %v",args)}
 }
}
