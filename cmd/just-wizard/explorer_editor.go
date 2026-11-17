//go:build desktop

package main

import (
 "bytes"
 "encoding/json"
 "os"
 "path/filepath"
 "strings"

 "fyne.io/fyne/v2"
 "fyne.io/fyne/v2/container"
 "fyne.io/fyne/v2/dialog"
 "fyne.io/fyne/v2/theme"
 "fyne.io/fyne/v2/widget"
)

func (u *desktopUI) textEditor(path string, data []byte, open func(), checkPath func()error)(fyne.CanvasObject,func()bool) {
 original:=append([]byte(nil),data...)
 entry:=widget.NewMultiLineEntry()
 entry.TextStyle=fyne.TextStyle{Monospace:true}
 entry.SetText(string(data))
 status:=widget.NewLabel("Saved")
 dirty:=func()bool{return !bytes.Equal(original,[]byte(entry.Text))}
 save:=widget.NewButtonWithIcon("Save",theme.DocumentSaveIcon(),func(){
  if !u.guard(){return}
  if err:=checkPath();err!=nil{dialog.ShowError(err,u.win);return}
  if err:=saveExplorerText(path,original,[]byte(entry.Text));err!=nil{dialog.ShowError(err,u.win);return}
  original=[]byte(entry.Text)
  status.SetText("Saved")
 })
 save.Importance=widget.HighImportance
 reload:=widget.NewButton("Reload",func(){
  action:=func(){
   if err:=checkPath();err!=nil{dialog.ShowError(err,u.win);return}
   info,err:=os.Stat(path);if err!=nil{dialog.ShowError(err,u.win);return}
   if info.Size()>previewLimit{dialog.ShowInformation("File too large","Reopen the file for read-only preview.",u.win);return}
   data,err:=os.ReadFile(path);if err!=nil{dialog.ShowError(err,u.win);return}
   original=data;entry.SetText(string(data));status.SetText("Saved")
  }
  if dirty(){dialog.ShowConfirm("Reload","Discard unsaved changes?",func(ok bool){if ok{action()}},u.win)}else{action()}
 })
 ext:=strings.ToLower(filepath.Ext(path))
 controls:=container.NewHBox(save,reload,widget.NewButton("Open externally",open))
 var content fyne.CanvasObject=entry
 var refreshPreview func(string)
 if ext==".md"||ext==".markdown"||ext==".mdown" {
  // Use Fyne's native Markdown renderer; source remains the canonical file.
  rich:=widget.NewRichTextFromMarkdown(entry.Text);rich.Wrapping=fyne.TextWrapWord
  refreshPreview=func(s string){rich.ParseMarkdown(s)}
  insert:=func(prefix,suffix,placeholder string){
   selected:=entry.SelectedText()
   if selected==""{selected=placeholder}
   // TypedRune uses the entry selection/caret rather than replacing unrelated matches.
   for _,r:=range prefix+selected+suffix{entry.TypedRune(r)}
  }
  formatting:=container.NewHBox(
   widget.NewButton("Bold",func(){insert("**","**","bold")}),
   widget.NewButton("Italic",func(){insert("*","*","italic")}),
   widget.NewButton("Heading",func(){insert("\n## ","\n","Heading")}),
   widget.NewButton("List",func(){insert("\n- ","\n","Item")}),
   widget.NewButton("Code",func(){insert("`","`","code")}),
  )
  split:=container.NewHSplit(entry,container.NewVScroll(container.NewPadded(rich)));split.Offset=0.5
  content=container.NewBorder(formatting,nil,nil,nil,split)
 }else if ext==".json"{
  controls.Add(widget.NewButton("Format JSON",func(){
   var out bytes.Buffer
   if err:=json.Indent(&out,[]byte(entry.Text),"","  ");err!=nil{dialog.ShowError(err,u.win);return}
   entry.SetText(out.String()+"\n")
  }))
 }
 controls.Add(widget.NewButton("Validate",func(){
  if err:=validateExplorerText(path,[]byte(entry.Text));err!=nil{dialog.ShowError(err,u.win);return}
  dialog.ShowInformation("Validation","No syntax errors detected by the available validator.",u.win)
 }))
 entry.OnChanged=func(s string){
  if dirty(){status.SetText("Unsaved changes")}else{status.SetText("Saved")}
  if refreshPreview!=nil{refreshPreview(s)}
 }
 return container.NewBorder(controls,status,nil,nil,content),dirty
}
