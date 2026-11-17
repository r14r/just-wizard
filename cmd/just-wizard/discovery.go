package main
import (
 "context"
 "encoding/json"
 "fmt"
 "os"
 "os/exec"
 "path/filepath"
 "sort"
 "strings"
 "time"
)
type parameter struct {
 Name string `json:"name"`
 Kind string `json:"kind"`
 Default json.RawMessage `json:"default"`
 Help string `json:"help"`
 Long string `json:"long"`
 Short string `json:"short"`
 Flag bool `json:"flag"`
 UI uiParam `json:"-"`
}
func(p parameter) optional()bool{return p.Flag||p.Kind=="star"||(len(p.Default)>0&&string(p.Default)!="null")}
func(p parameter) hint()string{
 s:="Required";if p.optional(){s="Optional — blank uses Just default"}
 if p.Kind=="plus"||p.Kind=="star"{s+="; space-separated values"}
 if p.Flag{s="true to enable; blank to omit"}
 if p.Help!=""{s=p.Help+" · "+s};if p.UI.Description!=""{s=p.UI.Description+" · "+s};if len(p.UI.Options)>0{s+="; allowed: "+strings.Join(p.UI.Options,", ")};return s
}
type recipe struct{Name string; Description string; Group string; Params []parameter; Label string; Confirm bool; Icon string}
type rawRecipe struct{
 Doc string `json:"doc"`
 Private bool `json:"private"`
 Attributes []json.RawMessage `json:"attributes"`
 Parameters []parameter `json:"parameters"`
}
type rawAlias struct{Target string `json:"target"`;Attributes []json.RawMessage `json:"attributes"`}
type rawFile struct{
 Recipes map[string]rawRecipe `json:"recipes"`
 Modules map[string]rawFile `json:"modules"`
 Aliases map[string]rawAlias `json:"aliases"`
}
func attributes(a []json.RawMessage)(groups []string,private bool){
 for _,v:=range a{
  var s string;if json.Unmarshal(v,&s)==nil{if s=="private"{private=true};continue}
  var m map[string]json.RawMessage;if json.Unmarshal(v,&m)!=nil{continue}
  if _,ok:=m["private"];ok{private=true}
  if g,ok:=m["group"];ok{var name string;if json.Unmarshal(g,&name)==nil&&name!=""{groups=append(groups,name)}}
 };return
}
func decodeRecipes(data []byte)([]recipe,error){
 var file rawFile;if err:=json.Unmarshal(data,&file);err!=nil{return nil,err}
 if file.Recipes==nil{return nil,fmt.Errorf("unsupported Just JSON dump: missing recipes")}
 var result []recipe
 var walk func(rawFile,string)
 walk=func(f rawFile,prefix string){
  add:=func(name string,r rawRecipe,attrs []json.RawMessage,desc string){
   gs,priv:=attributes(attrs);if priv||strings.HasPrefix(name,"_"){return}
   if len(gs)==0{gs=[]string{"Ungrouped"}}
   for _,g:=range gs{if prefix!=""{g=strings.TrimSuffix(prefix,"::")+" / "+g};result=append(result,recipe{Name:prefix+name,Description:desc,Group:g,Params:r.Parameters})}
  }
  for name,r:=range f.Recipes{if !r.Private{add(name,r,r.Attributes,r.Doc)}}
  for name,a:=range f.Aliases{if r,ok:=f.Recipes[a.Target];ok{attrs:=a.Attributes;if gs,_:=attributes(attrs);len(gs)==0{attrs=append(append([]json.RawMessage{},attrs...),r.Attributes...)};add(name,r,attrs,"Alias for "+a.Target+" — "+r.Doc)}}
  for name,m:=range f.Modules{if !strings.HasPrefix(name,"_"){walk(m,prefix+name+"::")}}
 };walk(file,"")
 sort.Slice(result,func(i,j int)bool{if result[i].Group==result[j].Group{return result[i].Name<result[j].Name};return result[i].Group<result[j].Group})
 return result,nil
}
func recipeGroups(rs []recipe)[]string{m:=map[string]bool{};var out []string;for _,r:=range rs{if !m[r.Group]{out=append(out,r.Group);m[r.Group]=true}};sort.Strings(out);return out}
func loadProject(root,file string)(string,string,[]recipe,error){
 abs,err:=filepath.Abs(root);if err!=nil{return "","",nil,err}
 if file==""{entries,e:=os.ReadDir(abs);if e!=nil{return "","",nil,e};for _,n:=range []string{"justfile",".justfile"}{for _,entry:=range entries{if !entry.IsDir()&&strings.EqualFold(entry.Name(),n){file=entry.Name();break}};if file!=""{break}};if file==""{return "","",nil,fmt.Errorf("no justfile in %s",abs)}}
 if !filepath.IsAbs(file){file=filepath.Join(abs,file)}
 ctx,cancel:=context.WithTimeout(context.Background(),15*time.Second);defer cancel()
 cmd:=exec.CommandContext(ctx,"just","--justfile",file,"--working-directory",abs,"--dump","--dump-format","json");cmd.Dir=abs
 out,err:=cmd.Output();if err!=nil{if e,ok:=err.(*exec.ExitError);ok{return "","",nil,fmt.Errorf("Just discovery: %s",e.Stderr)};return "","",nil,err}
 rs,err:=decodeRecipes(out);if err!=nil{return abs,file,nil,err};data,err:=os.ReadFile(file);if err!=nil{return abs,file,nil,err};return abs,file,applyMetadata(rs,data),nil
}
func recipeArgs(r recipe,values []string)([]string,error){
 if len(values)!=len(r.Params){return nil,fmt.Errorf("parameter count mismatch")}
 args:=[]string{r.Name};omitted:=false
 for i,p:=range r.Params{
  v:=values[i];if v==""&&p.UI.HasDefault{v=p.UI.Default};if err:=p.validate(v);err!=nil{return nil,err};option:="";if p.Long!=""{option="--"+p.Long}else if p.Short!=""{option="-"+p.Short}
  if p.Flag{if v=="true"&&option!=""{args=append(args,option)}else if v!=""&&v!="false"{return nil,fmt.Errorf("%s: enter true or false",p.Name)};continue}
  if v==""{if !p.optional(){return nil,fmt.Errorf("%s is required",p.Name)};if option==""{omitted=true};continue}
  if option!=""{args=append(args,option,v);continue}
  if omitted{return nil,fmt.Errorf("Supply earlier optional positional parameters before %s",p.Name)}
  if p.Kind=="plus"||p.Kind=="star"{args=append(args,strings.Fields(v)...)}else{args=append(args,v)}
 };return args,nil
}
func commandPreview(args []string)string{q:=[]string{"just"};for _,a:=range args{if strings.ContainsAny(a," \t\n'\"")||a==""{a="'"+strings.ReplaceAll(a,"'","'\\''")+"'"};q=append(q,a)};return "$ "+strings.Join(q," ")}
func risk(name string)string{for _,term:=range []string{"delete","remove","clean","reset","deploy","publish","release","push","uninstall","regenerate","rewrite","migrate"}{if strings.Contains(strings.ToLower(name),term){return "This recipe may modify files or external systems. Review before running."}};return ""}
