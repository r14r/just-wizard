package main

import (
 "os"
 "path/filepath"
 "reflect"
 "runtime"
 "strings"
 "testing"
)

func TestMetadataSourceAndUIOnlyFlag(t *testing.T) {
 source:=[]byte("# @ui param book type=select source=options-books-by-nr\n"+
  "# @ui param selected_engine type=select label=Engine source=options-engines flag=--engine\n"+
  "# @ui param model type=select source=options-models flag=--model\n"+
  "[positional-arguments]\nregenerate-images-book book *options:\n    @echo ok\n")
 r:=applyMetadata([]recipe{{Name:"regenerate-images-book",Params:[]parameter{
  {Name:"book"},{Name:"options",Kind:"star"},
 }}},source)[0]
 if len(r.Params)!=4{t.Fatalf("expected 4 parameters, got %d: %+v",len(r.Params),r.Params)}
 if r.Params[0].UI.Source!="options-books-by-nr"{t.Fatal("book source not parsed")}
 values:=[]string{"26","","sd15","local"} // UI-only params are alphabetical.
 if r.Params[2].Name!="model"||r.Params[3].Name!="selected_engine"{
  t.Fatalf("unexpected synthetic ordering: %+v",r.Params)
 }
 args,err:=recipeArgs(r,values)
 if err!=nil{t.Fatal(err)}
 want:=[]string{"regenerate-images-book","26","--model=sd15","--engine=local"}
 if !reflect.DeepEqual(args,want){t.Fatalf("got %v want %v",args,want)}
}

func TestStaticSelectCompatibility(t *testing.T) {
 src:=[]byte("# @ui param target type=select options=dev|prod\nbuild target:\n    @echo ok\n")
 r:=applyMetadata([]recipe{{Name:"build",Params:[]parameter{{Name:"target"}}}},src)[0]
 if !reflect.DeepEqual(r.Params[0].UI.Options,[]string{"dev","prod"}){t.Fatalf("%+v",r.Params[0])}
 if _,err:=recipeArgs(r,[]string{"invalid"});err==nil{t.Fatal("invalid value accepted")}
}

func TestUnknownAndConflictingSources(t *testing.T) {
 r:=recipe{Name:"build",Params:[]parameter{{Name:"book",UI:uiParam{Source:"options-books-by-nr"}}}}
 if _,err:=resolveParameterSources(r,".","justfile",[]recipe{r});err==nil{t.Fatal("missing source accepted")}
 for _,source:=range []string{"build","evil-command","options-../../etc","options-bad;touch"}{
  r.Params[0].UI.Source=source
  if _,err:=resolveParameterSources(r,".","justfile",[]recipe{r});err==nil{t.Errorf("accepted %s",source)}
 }
 r.Params[0].UI=uiParam{Source:"options-books-by-nr",Options:[]string{"1"}}
 options:=[]recipe{{Name:"options-books-by-nr"}}
 if _,err:=resolveParameterSources(r,".","justfile",options);err==nil{t.Fatal("source and options allowed together")}
}

func TestDynamicSourceExecution(t *testing.T) {
 if runtime.GOOS=="windows"{t.Skip("uses POSIX shell stub")}
 dir:=t.TempDir()
 justfile:=filepath.Join(dir,"project.just")
 if err:=os.WriteFile(justfile,[]byte("default:\n    @echo hello\n"),0600);err!=nil{t.Fatal(err)}
 fake:=filepath.Join(dir,"just")
 script:="#!/bin/sh\n"+
  "if [ \"$1\" != \"--justfile\" ] || [ \"$2\" != \""+justfile+"\" ]; then exit 8; fi\n"+
  "if [ \"$3\" != \"--working-directory\" ] || [ \"$4\" != \""+dir+"\" ]; then exit 9; fi\n"+
  "printf '26\\n27\\n27\\n'\n"
 if err:=os.WriteFile(fake,[]byte(script),0755);err!=nil{t.Fatal(err)}
 t.Setenv("PATH",dir+string(os.PathListSeparator)+os.Getenv("PATH"))
 r:=recipe{Name:"build",Params:[]parameter{{Name:"book",UI:uiParam{Source:"options-books-by-nr",Type:"select"}}}}
 resolved,err:=resolveParameterSources(r,dir,justfile,[]recipe{r,{Name:"options-books-by-nr"}})
 if err!=nil{t.Fatal(err)}
 if !reflect.DeepEqual(resolved.Params[0].UI.Options,[]string{"26","27"}){t.Fatalf("unexpected values %+v",resolved.Params[0].UI.Options)}
 if _,err=recipeArgs(resolved,[]string{"99"});err==nil{t.Fatal("unlisted option accepted")}
 if !strings.Contains(strings.Join(resolved.Params[0].UI.Options,","),"26"){t.Fatal("missing selected value")}
}

func TestDynamicSourceErrors(t *testing.T) {
 if runtime.GOOS=="windows"{t.Skip("uses POSIX shell stub")}
 dir:=t.TempDir()
 fake:=filepath.Join(dir,"just")
 t.Setenv("PATH",dir+string(os.PathListSeparator)+os.Getenv("PATH"))
 for _,tc:=range []struct{name,script string}{
  {"empty","#!/bin/sh\nexit 0\n"},
  {"error","#!/bin/sh\necho failed 1>&2\nexit 3\n"},
 }{
  t.Run(tc.name,func(t *testing.T){
   if err:=os.WriteFile(fake,[]byte(tc.script),0755);err!=nil{t.Fatal(err)}
   if _,err:=runSourceRecipe(dir,"justfile","options-books-by-nr");err==nil{t.Fatal("expected error")}
  })
 }
}
