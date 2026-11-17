//go:build desktop

package main

import (
 "bytes"
 "fmt"
 "io"
 "os"
 "os/exec"
 "path/filepath"
 "runtime"
 "strings"
 "unicode/utf8"

 "fyne.io/fyne/v2"
 "fyne.io/fyne/v2/canvas"
 "fyne.io/fyne/v2/container"
 "fyne.io/fyne/v2/dialog"
 "fyne.io/fyne/v2/theme"
 "fyne.io/fyne/v2/widget"
)

const previewLimit = 1024 * 1024

// explorerNode supplies a secondary-click target inside each tree row.
type explorerNode struct {
 widget.Label
 menu func(fyne.Position)
}
func newExplorerNode() *explorerNode {
 n:=&explorerNode{}
 n.ExtendBaseWidget(n)
 return n
}
func (n *explorerNode) TappedSecondary(e *fyne.PointEvent) {
 if n.menu!=nil {n.menu(e.AbsolutePosition)}
}

func (u *desktopUI) explorerWorkspace() (fyne.CanvasObject, fyne.CanvasObject) {
 root,err:=filepath.EvalSymlinks(u.root)
 if err!=nil{return widget.NewLabel("Project unavailable"),widget.NewLabel(err.Error())}
 // Do not traverse symlinks or paths outside the project root.
 resolve:=func(id string)(string,error){
  path:=filepath.Join(root,filepath.FromSlash(id))
  real,err:=filepath.EvalSymlinks(path);if err!=nil{return "",err}
  rel,err:=filepath.Rel(root,real)
  if err!=nil||rel==".."||strings.HasPrefix(rel,".."+string(os.PathSeparator)){return "",fmt.Errorf("Path is outside the project")}
  if real!=path{return "",fmt.Errorf("Symlink traversal is disabled")}
  return real,nil
 }
 var dirty func() bool
 preview:=container.NewStack(widget.NewLabel("Select a file to view its contents."))
 title:=widget.NewLabel("No file selected")
 title.Wrapping=fyne.TextWrapWord
 show:=func(obj fyne.CanvasObject){preview.Objects=[]fyne.CanvasObject{obj};preview.Refresh()}
 open:=func(id string){
  path,err:=resolve(id);if err!=nil{dialog.ShowError(err,u.win);return}
  var cmd *exec.Cmd
  switch runtime.GOOS {
  case "darwin":cmd=exec.Command("open",path)
  case "windows":cmd=exec.Command("rundll32","url.dll,FileProtocolHandler",path)
  default:cmd=exec.Command("xdg-open",path)
  }
  go func(){if err:=cmd.Run();err!=nil{fyne.Do(func(){dialog.ShowError(err,u.win)})}}()
 }
 var view func(string)
 view=func(id string){
  if dirty!=nil&&dirty(){dialog.ShowConfirm("Unsaved changes","Discard changes and open another file?",func(ok bool){if ok{dirty=nil;view(id)}},u.win);return}
  dirty=nil
  path,err:=resolve(id);if err!=nil{dialog.ShowError(err,u.win);return}
  info,err:=os.Stat(path);if err!=nil{dialog.ShowError(err,u.win);return}
  title.SetText(fmt.Sprintf("%s  ·  %d bytes",id,info.Size()))
  if info.IsDir(){show(widget.NewLabel("Expand the folder to select a file."));return}
  if !info.Mode().IsRegular(){show(widget.NewLabel("This is not a regular file."));return}
  ext:=strings.ToLower(filepath.Ext(path))
  switch ext{
  case ".png",".jpg",".jpeg",".gif",".webp",".svg":
   if info.Size()>20*1024*1024{show(widget.NewLabel("Image is too large for inline preview. Use Open."));return}
   image:=canvas.NewImageFromFile(path);image.FillMode=canvas.ImageFillContain
   show(image);return
  case ".pdf":
   show(u.documentPreview(path,false,func(){open(id)}));return
  case ".html",".htm":
   show(u.documentPreview(path,true,func(){open(id)}));return
  case ".epub",".zip",".mp4",".mov":
   show(container.NewCenter(container.NewVBox(widget.NewLabel("This format requires an external viewer."),widget.NewButton("Open file",func(){open(id)}))));return
  }
  f,err:=os.Open(path);if err!=nil{dialog.ShowError(err,u.win);return};defer f.Close()
  data,err:=io.ReadAll(io.LimitReader(f,previewLimit+1));if err!=nil{dialog.ShowError(err,u.win);return}
  truncated:=len(data)>previewLimit
  if truncated{
   data=data[:previewLimit]
   // A byte limit may cut through the final UTF-8 character.
   for i:=0;i<3&&!utf8.Valid(data);i++{data=data[:len(data)-1]}
  }
  if !utf8.Valid(data)||bytes.IndexByte(data,0)>=0{
   show(container.NewCenter(widget.NewLabel("Binary file. Use Open to view it externally.")));return
  }
  if truncated{
   title.SetText(title.Text+" · preview limited to 1 MiB (read-only)")
   text:=widget.NewMultiLineEntry();text.TextStyle=fyne.TextStyle{Monospace:true};text.SetText(string(data));text.Disable()
   show(text);return
  }
  editor,isDirty:=u.textEditor(path,data,func(){open(id)},func()error{_,err:=resolve(id);return err})
  dirty=isDirty
  show(editor)
 }
 children:=func(id string)[]string{
  path,err:=resolve(id);if err!=nil{return nil}
  entries,err:=os.ReadDir(path);if err!=nil{return nil}
  var dirs,files []string
  for _,entry:=range entries{
   name:=entry.Name()
   // Suppress repository internals, environments, symlinks and credential files.
   if name==".git"||name==".env"||strings.HasPrefix(name,".env.")||strings.HasPrefix(name,".venv")||name=="node_modules"||entry.Type()&os.ModeSymlink!=0{continue}
   child:=filepath.ToSlash(filepath.Join(id,name))
   if entry.IsDir(){dirs=append(dirs,child)}else if entry.Type().IsRegular(){files=append(files,child)}
  }
  return append(dirs,files...)
 }
 branch:=func(id string)bool{path,err:=resolve(id);if err!=nil{return false};info,err:=os.Stat(path);return err==nil&&info.IsDir()}
 tree:=widget.NewTree(children,branch,func(bool)fyne.CanvasObject{return newExplorerNode()},func(id string,isBranch bool,obj fyne.CanvasObject){
  node:=obj.(*explorerNode);node.Alignment=fyne.TextAlignLeading;node.SetText(filepath.Base(id))
  node.menu=nil
  if !isBranch{
   node.menu=func(pos fyne.Position){
    menu:=fyne.NewMenu("",fyne.NewMenuItem("Open",func(){open(id)}),fyne.NewMenuItem("View",func(){view(id)}))
    widget.ShowPopUpMenuAtPosition(menu,u.win.Canvas(),pos)
   }
  }
 })
 tree.OnSelected=func(id string){if !branch(id){view(id)}}
 // Keep the tree in the sidebar and the preview in the main workspace.
 sidebar:=container.NewBorder(container.NewBorder(nil,nil,heading("Project files"),widget.NewButtonWithIcon("",theme.ViewRefreshIcon(),func(){tree.Refresh()})),nil,nil,nil,tree)
 workspace:=container.NewBorder(title,nil,nil,nil,preview)
 u.explorerDirty=func()bool{return dirty!=nil&&dirty()}
 return sidebar,workspace
}
