//go:build desktop

package main

import (
 "context"
 "flag"
 "fmt"
 "image/color"
 "regexp"
 "os"
 "os/exec"
 "path/filepath"
 "sort"
 "strings"

 "fyne.io/fyne/v2"
 "fyne.io/fyne/v2/app"
 "fyne.io/fyne/v2/container"
 "fyne.io/fyne/v2/canvas"
 "fyne.io/fyne/v2/dialog"
 "fyne.io/fyne/v2/layout"
 "fyne.io/fyne/v2/widget"
 "fyne.io/fyne/v2/theme"
)

type desktopUI struct {
 win fyne.Window
 root string
 justfile string
 recipes []recipe
 body *fyne.Container
 log *fyne.Container
 logText string
 logScroll *container.Scroll
 status *widget.Label
 stop *widget.Button
 cancel context.CancelFunc
 busy bool
 explorerDirty func()bool
}

func heading(s string) *widget.Label { return widget.NewLabelWithStyle(s,fyne.TextAlignLeading,fyne.TextStyle{Bold:true}) }
func (u *desktopUI) show(content fyne.CanvasObject) {
 if u.explorerDirty!=nil&&u.explorerDirty(){dialog.ShowConfirm("Unsaved changes","Discard changes and leave the editor?",func(ok bool){if ok{u.explorerDirty=nil;u.show(content)}},u.win);return}
 u.explorerDirty=nil
 u.body.Objects=[]fyne.CanvasObject{content}
 u.body.Refresh()
}
func (u *desktopUI) guard() bool {
 if u.busy { dialog.ShowInformation("Task running","Wait for the current task or cancel it.",u.win);return false }
 return true
}
func (u *desktopUI) run(commands [][]string, finished func(error)) {
 if !u.guard() {return}
 if len(commands)==0{return}
 ctx,cancel:=context.WithCancel(context.Background())
 u.cancel=cancel;u.busy=true;u.stop.Enable();u.status.SetText("Running…")
 u.setOutput("")
 go func(){
  var result error
  var output strings.Builder
  for _,args:=range commands {
   if ctx.Err()!=nil {result=ctx.Err();break}
   output.WriteString(commandPreview(args)+"\n")
   snapshot:=output.String()
   fyne.Do(func(){u.setOutput(snapshot)})
   cmd:=exec.CommandContext(ctx,"just",append([]string{"--justfile",u.justfile,"--working-directory",u.root,"--"},args...)...)
   cmd.Dir=u.root
   data,err:=cmd.CombinedOutput()
   output.Write(data);output.WriteString("\n")
   if err!=nil {result=err;break}
  }
  if ctx.Err()!=nil{result=ctx.Err()}
  final:=output.String()
  if len(final)>120000 {final="[Earlier output omitted]\n"+final[len(final)-120000:]}
  fyne.Do(func(){
   u.setOutput(final)
   u.busy=false;u.cancel=nil;u.stop.Disable();cancel()
   if result!=nil {u.status.SetText("Failed / cancelled: "+result.Error())}else{u.status.SetText("Completed")}
   if finished!=nil{finished(result)}
  })
 }()
}
func (u *desktopUI) confirmRun(commands [][]string, finished func(error)) {
 var warnings []string
 for _,args:=range commands {if len(args)>0 {if reason:=risk(args[0]);reason!=""{warnings=append(warnings,args[0]+": "+reason)}}}
 if len(warnings)>0 {
  dialog.ShowConfirm("Confirm side effects",strings.Join(warnings,"\n"),func(ok bool){if ok{u.run(commands,finished)}},u.win)
 }else{u.run(commands,finished)}
}
// Sidebar tree: root topics and their individual recipe titles.
func (u *desktopUI) navigation() fyne.CanvasObject {
 children:=map[string][]string{"": {}}
 commands:=map[string]recipe{}
 for _,g:=range recipeGroups(u.recipes) {
  children[""]=append(children[""],g)
  children[g]=[]string{}
  var list []recipe
  for _,r:=range u.recipes {if r.Group==g{list=append(list,r)}}
  sort.Slice(list,func(i,j int)bool{return list[i].Name<list[j].Name})
  for _,r:=range list {id:=g+"/"+r.Name;children[g]=append(children[g],id);commands[id]=r}
 }
 tree:=widget.NewTree(
  func(id string)[]string{return children[id]},
  func(id string)bool{_,ok:=children[id];return ok},
  func(branch bool)fyne.CanvasObject{return widget.NewLabel("")},
  func(id string,branch bool,obj fyne.CanvasObject){
   label:=obj.(*widget.Label)
   label.Alignment=fyne.TextAlignLeading
   label.TextStyle=fyne.TextStyle{Bold:branch}
   if r,ok:=commands[id];ok{label.SetText(r.Name)}else{label.SetText(id)}
  },
 )
 // Accordion navigation: opening a topic closes every other topic.
 // The branch callback covers both the disclosure arrow and topic selection.
 tree.OnBranchOpened=func(id string){
  if id==""{return}
  for _,other:=range children[""]{
   if other!=id&&tree.IsBranch(other)&&tree.IsBranchOpen(other){tree.CloseBranch(other)}
  }
 }
 tree.OnSelected=func(id string){
  if r,ok:=commands[id];ok{u.commandDetail(r);return}
  if tree.IsBranchOpen(id){tree.CloseBranch(id)}else{tree.OpenBranch(id)}
 }
 return tree
}

func (u *desktopUI) commandDetail(r recipe) {
 description:=r.Description
 if description=="" {description="Execute the "+r.Name+" recipe from the project Justfile."}
 desc:=widget.NewLabel(description);desc.Wrapping=fyne.TextWrapWord
 fields:=make([]*widget.Entry,len(r.Params))
 items:=make([]*widget.FormItem,0,len(fields))
 for i,p:=range r.Params {
  e:=widget.NewEntry()
  
  e.SetPlaceHolder(p.hint())
  fields[i]=e;items=append(items,widget.NewFormItem(p.Name,e))
 }
 args:=func()([]string,error){
  values:=make([]string,len(fields))
  for i,e:=range fields{values[i]=e.Text}
  return recipeArgs(r,values)
 }
 preview:=widget.NewLabel("");preview.TextStyle=fyne.TextStyle{Monospace:true};preview.Wrapping=fyne.TextWrapWord
 update:=func(){a,err:=args();if err!=nil{preview.SetText(err.Error());return};preview.SetText(commandPreview(a))}
 for _,e:=range fields{e.OnChanged=func(string){update()}}
 update()
 var parameters fyne.CanvasObject=widget.NewLabel("No parameters required.")
 if len(items)>0 {parameters=widget.NewForm(items...)}
 details:=container.New(layout.NewFormLayout(),
  container.NewVBox(heading("Description:")),desc,
  container.NewVBox(heading("Parameters:")),parameters,
  container.NewVBox(heading("Command:")),preview,
 )
 content:=container.NewVBox(heading(r.Group+" / "+r.Name),widget.NewSeparator(),details,widget.NewSeparator())
 if reason:=risk(r.Name);reason!=""{warning:=widget.NewLabel("Warning: "+reason);warning.Wrapping=fyne.TextWrapWord;content.Add(warning)}
 run:=widget.NewButtonWithIcon("Run",theme.MediaPlayIcon(),func(){if u.guard(){a,err:=args();if err!=nil{dialog.ShowError(err,u.win);return};u.confirmRun([][]string{a},nil)}})
 run.Importance=widget.HighImportance
 content.Add(container.NewHBox(run,widget.NewButton("Reset parameters",func(){for _,e:=range fields{e.SetText("")};update()})))
 u.show(container.NewVScroll(container.NewPadded(content)))
}

var ansiOutput=regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)
func (u *desktopUI) setOutput(s string){
 u.logText=ansiOutput.ReplaceAllString(s,"")
 u.log.Objects=nil
 for _,line:=range strings.Split(u.logText,"\n"){
  if line==""{line=" "}
  row:=canvas.NewText(strings.ReplaceAll(line,"\t","    "),color.White)
  row.TextStyle=fyne.TextStyle{Monospace:true};row.TextSize=14
  u.log.Objects=append(u.log.Objects,row)
 }
 u.log.Refresh()
 if u.logScroll!=nil{u.logScroll.Refresh();u.logScroll.ScrollToBottom()}
}
func main(){
 root:=flag.String("root",".","project directory")
 file:=flag.String("justfile","","explicit justfile (relative to root or absolute)")
 flag.Parse()
 abs,jf,recipes,err:=loadProject(*root,*file)
 if err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(1)}
 a:=app.NewWithID("de.digitalnexus.justwizard")
 w:=a.NewWindow("Just Wizard — "+filepath.Base(abs))
 u:=&desktopUI{win:w,root:abs,justfile:jf,recipes:recipes,body:container.NewStack(),log:container.NewVBox(),status:widget.NewLabel("Ready")}
 u.stop=widget.NewButton("Cancel task",func(){if u.cancel!=nil{u.cancel()}});u.stop.Disable()
 runTab:=container.NewTabItem("RUN",u.navigation())
 explorerSidebar:=container.NewStack(widget.NewLabel("Select EXPLORER to browse project files."))
 explorerTab:=container.NewTabItem("EXPLORER",explorerSidebar)
 tabs:=container.NewAppTabs(runTab,explorerTab)
 active:=runTab
 var runContent []fyne.CanvasObject
 switching:=false
 switchMode:=func(target *container.TabItem){
  if target==explorerTab{
   runContent=append([]fyne.CanvasObject(nil),u.body.Objects...)
   sidebar,content:=u.explorerWorkspace()
   explorerSidebar.Objects=[]fyne.CanvasObject{sidebar};explorerSidebar.Refresh()
   u.body.Objects=[]fyne.CanvasObject{content}
  }else{
   u.explorerDirty=nil
   u.body.Objects=runContent
  }
  u.body.Refresh()
  active=target
 }
 tabs.OnSelected=func(target *container.TabItem){
  if switching||target==active{return}
  if active==explorerTab&&u.explorerDirty!=nil&&u.explorerDirty(){
   switching=true;tabs.Select(active);switching=false
   dialog.ShowConfirm("Unsaved changes","Discard changes and return to RUN?",func(ok bool){
    if ok{switching=true;tabs.Select(target);switching=false;switchMode(target)}
   },w)
   return
  }
  switchMode(target)
 }
 nav:=container.NewBorder(heading("JUST WIZARD"),nil,nil,nil,tabs)
 u.logScroll=container.NewScroll(container.NewPadded(u.log))
 blue:=canvas.NewRectangle(color.NRGBA{R:15,G:48,B:97,A:255})
 outputArea:=container.NewStack(blue,u.logScroll)
 output:=container.NewBorder(container.NewHBox(heading("Command output"),u.stop,widget.NewButton("Copy",func(){w.Clipboard().SetContent(u.logText)}),widget.NewButton("Clear",func(){u.setOutput("")})),u.status,nil,nil,outputArea)
 workspace:=container.NewVSplit(u.body,output);workspace.Offset=0.65
 mainSplit:=container.NewHSplit(nav,workspace);mainSplit.Offset=0.29
 w.SetContent(mainSplit);w.Resize(fyne.NewSize(1260,850))
 w.SetCloseIntercept(func(){if u.explorerDirty!=nil&&u.explorerDirty(){dialog.ShowConfirm("Unsaved changes","Discard unsaved file changes and quit?",func(ok bool){if ok{u.explorerDirty=nil;if u.cancel!=nil{u.cancel()};w.SetCloseIntercept(nil);w.Close()}},w);return};if u.busy{dialog.ShowConfirm("Quit","Cancel running task and close?",func(ok bool){if ok{if u.cancel!=nil{u.cancel()};w.SetCloseIntercept(nil);w.Close()}},w)}else{w.SetCloseIntercept(nil);w.Close()}})
 u.show(container.NewPadded(container.NewVBox(heading("Project: "+filepath.Base(abs)),widget.NewLabel(jf),widget.NewLabel(fmt.Sprintf("%d menu entries · select a recipe in RUN",len(recipes))))));w.ShowAndRun()
}
