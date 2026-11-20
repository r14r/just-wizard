package main

import (
 "bytes"
 "encoding/json"
 "flag"
 "fmt"
 "math"
 "os"
 "os/exec"
 "path/filepath"
 "regexp"
 "strconv"
 "strings"
)

type uiParam struct {
 Label, Description, Type, Default, Min, Max string
 HasDefault, Required bool
 Options []string
}
type uiMetadata struct {
 Group, Label, Description, Icon string
 Confirm bool
 Params map[string]uiParam
}
type sourceRecipe struct {
 start int
 meta uiMetadata
}
var sourceHeader = regexp.MustCompile("^(?:@)?([A-Za-z_][A-Za-z0-9_-]*)(?:[ \\t]+[^:]*)?:")
var metaToken = regexp.MustCompile("(?:[^\\s\"']|\"[^\"]*\"|'[^']*')+")

// Only simple root-level headers are annotated. The Just JSON dump remains
// authoritative for actual recipes, aliases, modules and executable parameters.
func sourceRecipes(lines []string) map[string]sourceRecipe {
 out:=map[string]sourceRecipe{}
 start:=-1
 for i,line:=range lines {
  line=strings.TrimSuffix(line,"\r")
  if strings.HasPrefix(line,"#") || strings.HasPrefix(line,"[") {
   if start<0 {start=i}; continue
  }
  if line=="" || strings.TrimSpace(line)=="" {start=-1;continue}
  m:=sourceHeader.FindStringSubmatch(line)
  if m!=nil && !strings.Contains(line,":=") {
   at:=start;if at<0{at=i}
   md:=uiMetadata{Params:map[string]uiParam{}}
   for _,comment:=range lines[at:i] {
    s:=strings.TrimSpace(strings.TrimPrefix(comment,"#"))
    if !strings.HasPrefix(s,"@ui "){continue}
    key,value,_:=strings.Cut(strings.TrimSpace(strings.TrimPrefix(s,"@ui "))," ")
    value=strings.TrimSpace(value)
    switch key {
    case "group":md.Group=value
    case "label":md.Label=value
    case "icon":md.Icon=value
    case "description":md.Description=value
    case "confirm":md.Confirm=true
    case "param":
     tokens:=metaToken.FindAllString(value,-1);if len(tokens)==0{continue}
     a:=md.Params[tokens[0]]
     for _,tok:=range tokens[1:] {
      k,v,ok:=strings.Cut(tok,"=");if !ok{if tok=="required"{a.Required=true};continue}
      v=strings.Trim(v,"\"'")
      switch k {
      case "label":a.Label=strings.ReplaceAll(v,"_"," ")
      case "description":a.Description=strings.ReplaceAll(v,"_"," ")
      case "type":a.Type=v
      case "default":a.Default=v;a.HasDefault=true
      case "min":a.Min=v
      case "max":a.Max=v
      case "options":a.Options=strings.Split(v,"|")
      }
     }
     md.Params[tokens[0]]=a
    }
   }
   out[m[1]]=sourceRecipe{at,md}
  }
  start=-1
 }
 return out
}
func applyMetadata(rs []recipe, data []byte) []recipe {
 src:=sourceRecipes(strings.Split(string(data),"\n"))
 for i:=range rs {
  r:=&rs[i]; s,ok:=src[r.Name];if !ok{continue};m:=s.meta
  if m.Group!=""{r.Group=m.Group};r.Label=m.Label;r.Icon=m.Icon
  if m.Description!=""{r.Description=m.Description};r.Confirm=m.Confirm
  r.Params=append([]parameter(nil),r.Params...)
  for j:=range r.Params{r.Params[j].UI=m.Params[r.Params[j].Name]}
 }
 // Recipes with multiple native groups collapse when explicit metadata overrides them.
 seen:=map[string]bool{};out:=rs[:0]
 for _,r:=range rs {key:=r.Group+"\x00"+r.Name;if !seen[key]{seen[key]=true;out=append(out,r)}}
 return out
}
func (r recipe) title()string{if r.Label!=""{return r.Label};return r.Name}
func (p parameter) label()string{if p.UI.Label!=""{return p.UI.Label};return p.Name}
func (p parameter) initial()string{
 if p.UI.HasDefault{return p.UI.Default}
 if p.Flag||p.UI.Type=="checkbox"{return "false"}
 var v string;if json.Unmarshal(p.Default,&v)==nil&&string(p.Default)!="null"{return v}
 return ""
}
func (p parameter) validate(v string)error{
 if v==""{if p.UI.Required{return fmt.Errorf("%s is required",p.Name)};return nil}
 if p.UI.Type=="select"||len(p.UI.Options)>0 {
  valid:=false;for _,o:=range p.UI.Options{if o==v{valid=true}}
  if len(p.UI.Options)>0&&!valid{return fmt.Errorf("%s: choose one of %s",p.Name,strings.Join(p.UI.Options,", "))}
 }
 if p.UI.Type=="checkbox"&&v!="true"&&v!="false"{return fmt.Errorf("%s: enter true or false",p.Name)}
 if p.UI.Type=="number"||p.UI.Min!=""||p.UI.Max!=""{
  n,e:=strconv.ParseFloat(v,64);if e!=nil||math.IsNaN(n)||math.IsInf(n,0){return fmt.Errorf("%s must be a finite number",p.Name)}
  for _,bound:=range []struct{value string;min bool}{{p.UI.Min,true},{p.UI.Max,false}}{
   if bound.value==""{continue};b,e:=strconv.ParseFloat(bound.value,64)
   if e!=nil||math.IsNaN(b)||math.IsInf(b,0){return fmt.Errorf("%s: invalid numeric constraint",p.Name)}
   if bound.min&&n<b||!bound.min&&n>b{return fmt.Errorf("%s must be within configured bounds",p.Name)}
  }
 }
 return nil
}
func inferredGroup(n string)string{
 switch strings.Split(n,"-")[0]{
 case "build","compile","make":return "Build"
 case "test","ci","check","lint","verify","validate","doctor":return "Development"
 case "clean","remove","delete","uninstall","migrate","repair":return "Maintenance"
 case "install","bootstrap","setup","init":return "Setup"
 case "wizard","terminal","run","serve","preview":return "Application"
 case "package","release","deploy","publish","export":return "Publishing"
 case "default","help","list","version","show":return "Information"
 };return "Other"
}
func inferredType(p parameter)string{
 if p.Flag{return "checkbox"}
 switch strings.ToLower(p.Name){case "root","path","file","directory","dir":return "path"}
 if strings.Contains(strings.ToLower(p.Name),"password"){return "password"}
 return "text"
}
func migrateSource(data []byte,rs []recipe)([]byte,int){
 newline:="\n";if bytes.Contains(data,[]byte("\r\n")){newline="\r\n"}
 lines:=strings.Split(string(data),newline);src:=sourceRecipes(lines)
 inserts:=map[int][]string{};seen:=map[string]bool{}
 for _,r:=range rs{
  s,ok:=src[r.Name];if !ok||seen[r.Name]{continue};seen[r.Name]=true;m:=s.meta
  var add []string
  g:=m.Group;if g==""{g=r.Group;if g==""||g=="Ungrouped"{g=inferredGroup(r.Name)};add=append(add,"# @ui group "+g)}
  if m.Icon==""{r.Group=g;add=append(add,"# @ui icon "+r.iconID())}
  if m.Label==""{label:=strings.ReplaceAll(strings.ReplaceAll(r.Name,"-"," "),"_"," ");if len(label)>0{label=strings.ToUpper(label[:1])+label[1:]};add=append(add,"# @ui label "+label)}
  if m.Description==""{d:=r.Description;if d==""||strings.Contains(d,"@ui "){d="Run the "+r.Name+" recipe."};d=strings.Join(strings.Fields(d)," ");add=append(add,"# @ui description "+d)}
  if !m.Confirm&&risk(r.Name)!=""{add=append(add,"# @ui confirm")}
  for _,p:=range r.Params{
   if _,ok:=m.Params[p.Name];ok{continue}
   add=append(add,"# @ui param "+p.Name+" type="+inferredType(p)+" label="+strings.ReplaceAll(p.Name,"-","_"))
  }
  if len(add)>0{inserts[s.start]=add}
 }
 var out []string
 for i,line:=range lines{out=append(out,inserts[i]...);out=append(out,line)}
 return []byte(strings.Join(out,newline)),len(inserts)
}
// Both desktop and terminal binaries dispatch migrate before launching their UI.
func dispatchMigration()bool{
 if len(os.Args)<2{return false}
 if os.Args[1]=="options"{if err:=runOptions(os.Args[2:]);err!=nil{fmt.Fprintln(os.Stderr,"options:",err);os.Exit(1)};return true}
 if os.Args[1]!="migrate"{return false}
 if err:=runMigration(os.Args[2:]);err!=nil{fmt.Fprintln(os.Stderr,"migrate:",err);os.Exit(1)}
 return true
}
func runMigration(args []string)error{
 fs:=flag.NewFlagSet("migrate",flag.ContinueOnError)
 root:=fs.String("root",".","project directory")
 file:=fs.String("justfile","","explicit justfile")
 dry:=fs.Bool("dry-run",false,"print proposed justfile without writing")
 if err:=fs.Parse(args);err!=nil{if err==flag.ErrHelp{return nil};return err}
 if fs.NArg()!=0{return fmt.Errorf("unexpected argument: %s",fs.Arg(0))}
 dir,path,rs,err:=loadProject(*root,*file);if err!=nil{return err}
 path,err=filepath.EvalSymlinks(path);if err!=nil{return err}
 old,err:=os.ReadFile(path);if err!=nil{return err}
 updated,n:=migrateSource(old,rs)
 if *dry{fmt.Print(string(updated));return nil}
 if n==0{fmt.Println("No migration needed:",path);return nil}
 info,err:=os.Stat(path);if err!=nil{return err}
 tmp,err:=os.CreateTemp(filepath.Dir(path),".just-wizard-migrate-*");if err!=nil{return err}
 defer os.Remove(tmp.Name())
 if err=tmp.Chmod(info.Mode().Perm());err==nil{_,err=tmp.Write(updated)}
 closeErr:=tmp.Close();if err!=nil{return err};if closeErr!=nil{return closeErr}
 cmd:=exec.Command("just","--justfile",tmp.Name(),"--working-directory",dir,"--dump","--dump-format","json");cmd.Dir=dir
 if out,e:=cmd.CombinedOutput();e!=nil{return fmt.Errorf("proposed justfile rejected: %w: %s",e,out)}
 current,err:=os.ReadFile(path);if err!=nil{return err};if !bytes.Equal(current,old){return fmt.Errorf("justfile changed during migration")}
 backup,err:=os.OpenFile(path+".bak",os.O_WRONLY|os.O_CREATE|os.O_EXCL,info.Mode().Perm());if err!=nil{return fmt.Errorf("cannot create backup (move existing .bak first): %w",err)}
 _,err=backup.Write(old);ce:=backup.Close();if err!=nil{return err};if ce!=nil{return ce}
 if err=os.Rename(tmp.Name(),path);err!=nil{return err}
 fmt.Printf("Migrated %d recipes in %s\nBackup: %s.bak\n",n,path,path)
 return nil
}

func safePreview(r recipe,args []string)string{
 for _,p:=range r.Params{if p.UI.Type=="password"{return "$ just "+r.Name+" [arguments hidden: password parameter]"}}
 return commandPreview(args)
}
