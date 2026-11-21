package main

import (
 "context"
 "fmt"
 "os/exec"
 "regexp"
 "strings"
 "time"
)

// Source recipes are deliberately limited to options-* commands. These are
// regular Just recipes in the current project and should be trusted just as
// ordinary project commands are. No shell is involved in command execution.
var sourceNamePattern=regexp.MustCompile(`^options-[a-z0-9][a-z0-9-]*$`)

func resolveParameterSources(r recipe,root,justfile string,all []recipe)(recipe,error){
 resolved:=r
 resolved.Params=append([]parameter(nil),r.Params...)
 known:=map[string]bool{}
 for _,item:=range all{known[item.Name]=true}
 for i:=range resolved.Params{
  p:=&resolved.Params[i]
  if p.UI.Source==""{continue}
  if len(p.UI.Options)>0{return r,fmt.Errorf("%s: use either source= or options=, not both",p.Name)}
  source:=p.UI.Source
  if !sourceNamePattern.MatchString(source)||!known[source]||source==r.Name{
   return r,fmt.Errorf("%s: invalid or unknown selector source %q",p.Name,source)
  }
  choices,err:=runSourceRecipe(root,justfile,source)
  if err!=nil{return r,fmt.Errorf("%s: %w",p.Name,err)}
  p.UI.Options=choices
 }
 return resolved,nil
}

func runSourceRecipe(root,justfile,source string)([]string,error){
 if !sourceNamePattern.MatchString(source){return nil,fmt.Errorf("invalid selector source %q",source)}
 ctx,cancel:=context.WithTimeout(context.Background(),8*time.Second)
 defer cancel()
 command:=exec.CommandContext(ctx,"just","--justfile",justfile,"--working-directory",root,"--quiet","--",source)
 command.Dir=root
 out,err:=command.Output()
 if ctx.Err()!=nil{return nil,fmt.Errorf("selector %s timed out: %w",source,ctx.Err())}
 if err!=nil{
  if e,ok:=err.(*exec.ExitError);ok{return nil,fmt.Errorf("selector %s failed: %w: %s",source,err,strings.TrimSpace(string(e.Stderr)))}
  return nil,fmt.Errorf("selector %s failed: %w",source,err)
 }
 if len(out)>65536{return nil,fmt.Errorf("selector %s output exceeds 64 KiB",source)}
 choices:=[]string{};seen:=map[string]bool{}
 for _,raw:=range strings.Split(string(out),"\n"){
  v:=strings.TrimSpace(raw)
  if v==""{continue}
  if strings.ContainsAny(v,"\x00\r\t"){return nil,fmt.Errorf("selector %s returned an invalid value",source)}
  if !seen[v]{seen[v]=true;choices=append(choices,v)}
  if len(choices)>500{return nil,fmt.Errorf("selector %s returned too many choices",source)}
 }
 if len(choices)==0{return nil,fmt.Errorf("selector %s returned no choices",source)}
 return choices,nil
}
