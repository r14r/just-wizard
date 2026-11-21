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
   display:=commandPreview(args);for _,r:=range u.recipes{if len(args)>0&&r.Name==args[0]{display=safePreview(r,args);break}};output.WriteString(display+"\n")
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
func (u *desktopUI) confirmRun(commands [][]string, started func(), finished func(error)) {
 var warnings []string
 for _,args:=range commands {if len(args)>0 {if reason:=risk(args[0]);reason!=""{warnings=append(warnings,args[0]+": "+reason)}}}
 if len(warnings)>0 {
  dialog.ShowConfirm("Confirm side effects",strings.Join(warnings,"\n"),func(ok bool){if ok{if started!=nil{started()};u.run(commands,finished)}},u.win)
 }else{if started!=nil{started()};u.run(commands,finished)}
}
// recipeIcon renders a canonical icon ID using existing themed resources.
func recipeIcon(id string) fyne.Resource {
 switch id {
 case "terminal":
  return sidebarSymbol("terminal", `<path d="M3 4h18v16H3z M6 8l4 4-4 4m7 0h5"/>`)
 case "package":
  return sidebarSymbol("package", `<path d="M12 3l9 5v8l-9 5-9-5V8z M3 8l9 5 9-5m-9 5v8M7.5 5.5l9 5"/>`)
 case "hammer":
  return sidebarSymbol("build", `<path d="M14 5l5 5m-2-7 4 4-5 5-2-2-9 9-2-2 9-9-2-2 3-3z"/>`)
 case "flask-conical":
  return sidebarSymbol("test", `<path d="M9 3h6m-5 0v7l-5 8a2 2 0 0 0 2 3h10a2 2 0 0 0 2-3l-5-8V3M8 15h8"/>`)
 case "clipboard-check":
  return sidebarSymbol("check", `<path d="M8 4H5v17h14V4h-3M8 3h8v4H8z M8 14l3 3 5-6"/>`)
 case "activity":
  return sidebarSymbol("health", `<path d="M3 12h4l3-7 4 14 3-7h4"/>`)
 case "trash-2":
  return sidebarSymbol("trash", `<path d="M3 6h18M9 6V3h6v3M5 6l1 15h12l1-15M10 10v7m4-7v7"/>`)
 case "download":
  return sidebarSymbol("download", `<path d="M12 3v12m-5-5 5 5 5-5M4 16v5h16v-5"/>`)
 case "upload":
  return sidebarSymbol("upload", `<path d="M12 16V4m-5 5 5-5 5 5M4 16v5h16v-5"/>`)
 case "tag":
  return sidebarSymbol("tag", `<path d="M3 3h8l10 10-8 8L3 11z M7 7h.1"/>`)
 case "monitor":
  return sidebarSymbol("monitor", `<path d="M3 4h18v13H3z M12 17v4m-5 0h10"/>`)
 case "list":
  return sidebarSymbol("list", `<path d="M8 6h13M8 12h13M8 18h13M3 6h.1M3 12h.1M3 18h.1"/>`)
 case "settings":
  return theme.SettingsIcon()
 case "square":
  return theme.MediaStopIcon()
 case "refresh-cw":
  return theme.ViewRefreshIcon()
 case "archive":
  return sidebarSymbol("archive", `<path d="M3 3h18v5H3z M5 8v13h14V8M9 12h6"/>`)
 default:
  return theme.MediaPlayIcon()
 }
}

// Embedded outline icons follow Fyne's foreground and selected-button theme.
func sidebarSymbol(name, shape string) fyne.Resource {
 svg := `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#000000" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">` + shape + `</svg>`
 return theme.NewThemedResource(fyne.NewStaticResource("recipe-"+name+".svg", []byte(svg)))
}

// Sidebar navigation: flat, icon-led rows with independently expandable groups.
func (u *desktopUI) navigation() fyne.CanvasObject {
 rows := container.NewVBox()
 expanded := map[string]bool{}
 selected := ""
 groups := recipeGroups(u.recipes)
 var redraw func()
 // Low-importance buttons retain native hover, keyboard focus and click handling.
 // The active recipe uses the theme's primary color.
 makeRow := func(label string, icon fyne.Resource, active bool, indent bool, action func()) fyne.CanvasObject {
  button := widget.NewButtonWithIcon(label, icon, action)
  button.Alignment = widget.ButtonAlignLeading
  button.Importance = widget.LowImportance
  if active { button.Importance = widget.HighImportance }
  if !indent { return button }
  space := canvas.NewRectangle(color.Transparent)
  space.SetMinSize(fyne.NewSize(24, 1))
  return container.NewBorder(nil, nil, space, nil, button)
 }
 groupIcon := func(group string) fyne.Resource {
  switch strings.ToLower(group) {
  case "development", "build": return theme.SettingsIcon()
  case "application", "run": return theme.MediaPlayIcon()
  default: return theme.FolderIcon()
  }
 }
 redraw = func() {
  rows.Objects = nil
  for _, group := range groups {
   g := group
   var list []recipe
   for _, r := range u.recipes { if r.Group == g { list = append(list, r) } }
   sort.Slice(list, func(i,j int) bool { return list[i].Name < list[j].Name })
   toggle := func() { expanded[g] = !expanded[g]; redraw() }
   arrow := theme.NavigateNextIcon()
   if expanded[g] { arrow = theme.MoveDownIcon() }
   disclosure := widget.NewButtonWithIcon("", arrow, toggle)
   disclosure.Importance = widget.LowImportance
   title := makeRow(g, groupIcon(g), false, false, toggle)
   rows.Add(container.NewBorder(nil, nil, nil, disclosure, title))
   if !expanded[g] { continue }
   for _, item := range list {
    r := item
    id := g + "\x00" + r.Name
    rows.Add(makeRow(r.title(), recipeIcon(r.iconID()), selected == id, true, func() {
     selected = id
     u.commandDetail(r)
     redraw()
    }))
   }
  }
  rows.Refresh()
 }
 if len(groups) > 0 { expanded[groups[0]] = true }
 redraw()
 border := canvas.NewRectangle(color.Transparent)
 border.StrokeColor = theme.DisabledColor()
 border.StrokeWidth = 1
 border.CornerRadius = 12
 return container.NewStack(border, container.NewPadded(container.NewVScroll(rows)))
}

func (u *desktopUI) commandDetail(r recipe) {
 resolved,err:=resolveParameterSources(r,u.root,u.justfile,u.recipes)
 if err!=nil{
  u.show(container.NewVBox(heading(r.title()),widget.NewLabel("Parameter source error: "+err.Error()),widget.NewButton("Retry",func(){u.commandDetail(r)})))
  return
 }
 r=resolved
 description:=r.Description
 if description=="" {description="Execute the "+r.Name+" recipe from the project Justfile."}
 desc:=widget.NewLabel(description);desc.Wrapping=fyne.TextWrapWord
 type formField struct{value func()string}
 fields:=make([]formField,0,len(r.Params))
 items:=make([]*widget.FormItem,0,len(r.Params))
 var update func()
 changed:=func(){if update!=nil{update()}}
 for _,p:=range r.Params {
  p:=p
  initial:=p.initial()
  var obj fyne.CanvasObject
  var field formField
  switch {
  case len(p.UI.Options)>0:
   sel:=widget.NewSelect(p.UI.Options,func(string){changed()})
   sel.SetSelected(initial)
   obj=sel;field=formField{func()string{return sel.Selected}}
  case p.UI.Type=="checkbox"||p.Flag:
   check:=widget.NewCheck("",func(bool){changed()});check.SetChecked(initial=="true")
   obj=check;field=formField{func()string{if check.Checked{return "true"};return "false"}}
  default:
   e:=widget.NewEntry();if p.UI.Type=="password"{e=widget.NewPasswordEntry()}
   e.SetPlaceHolder(p.hint());e.SetText(initial);e.OnChanged=func(string){changed()}
   obj=e;field=formField{func()string{return e.Text}}
  }
  if p.UI.Description!=""{note:=widget.NewLabel(p.UI.Description);note.Wrapping=fyne.TextWrapWord;obj=container.NewVBox(obj,note)}
  fields=append(fields,field);items=append(items,widget.NewFormItem(p.label(),obj))
 }
 args:=func()([]string,error){
  values:=make([]string,len(fields))
  for i,f:=range fields{values[i]=f.value()}
  return recipeArgs(r,values)
 }
 preview:=widget.NewLabel("");preview.TextStyle=fyne.TextStyle{Monospace:true};preview.Wrapping=fyne.TextWrapWord
 update=func(){a,err:=args();if err!=nil{preview.SetText(err.Error());return};preview.SetText(safePreview(r,a))}
 update()
 var parameters fyne.CanvasObject=widget.NewLabel("No parameters required.")
 if len(items)>0 {parameters=widget.NewForm(items...)}
 details:=container.New(layout.NewFormLayout(),
  container.NewVBox(heading("Description:")),desc,
  container.NewVBox(heading("Parameters:")),parameters,
  container.NewVBox(heading("Command:")),preview,
 )
 content:=container.NewVBox(heading(r.Group+" / "+r.title()),widget.NewSeparator(),details,widget.NewSeparator())
 if reason:=risk(r.Name);reason!=""{warning:=widget.NewLabel("Warning: "+reason);warning.Wrapping=fyne.TextWrapWord;content.Add(warning)}
 // Keep progress and result immediately beside Run, even for fast recipes.
 // A radial activity wheel with its caption below, matching the Building indicator.
 progress:=widget.NewActivity()
 buildingLabel:=widget.NewLabelWithStyle("Building",fyne.TextAlignCenter,fyne.TextStyle{Bold:true})
 building:=container.NewVBox(
  container.NewCenter(container.NewGridWrap(fyne.NewSize(52,52),progress)),
  buildingLabel,
 )
 building.Hide()
 result:=widget.NewLabel("")
 result.Hide()
 var run *widget.Button
 started:=func(){
  result.Hide()
  building.Show()
  progress.Start()
  run.Disable()
 }
 finished:=func(err error){
  progress.Stop()
  building.Hide()
  run.Enable()
  if err!=nil {result.SetText("Failed: "+err.Error())} else {result.SetText("✓ OK — Completed")}
  result.Show()
 }
 run=widget.NewButtonWithIcon("Run",theme.MediaPlayIcon(),func(){
  if !u.guard(){return}
  a,err:=args()
  if err!=nil{dialog.ShowError(err,u.win);return}
  execute:=func(){u.confirmRun([][]string{a},started,finished)}
  if r.Confirm&&risk(r.Name)==""{
   dialog.ShowConfirm("Confirm recipe","Run "+r.title()+"?",func(ok bool){if ok{execute()}},u.win)
  }else{execute()}
 })
 run.Importance=widget.HighImportance
 content.Add(container.NewHBox(run,building,result))
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
 if dispatchMigration(){return}
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
