//go:build desktop

package main

import (
 "context"
 "fmt"
 "os"
 "net/url"
 "os/exec"
 "path/filepath"
 "runtime"
 "strings"
 "time"

 "fyne.io/fyne/v2"
 "fyne.io/fyne/v2/canvas"
 "fyne.io/fyne/v2/container"
 "fyne.io/fyne/v2/widget"
)

func browserBinary() string {
 for _,name:=range []string{"chromium","chromium-browser","google-chrome","google-chrome-stable","chrome"}{if p,err:=exec.LookPath(name);err==nil{return p}}
 if runtime.GOOS=="darwin"{for _,p:=range []string{"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome","/Applications/Chromium.app/Contents/MacOS/Chromium","/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge"}{if _,err:=os.Stat(p);err==nil{return p}}}
 return ""
}

func (u *desktopUI) documentPreview(path string, html bool, open func()) fyne.CanvasObject {
 pane:=container.NewStack(widget.NewLabel("Preparing preview…"))
 status:=widget.NewLabel("")
 status.Wrapping=fyne.TextWrapWord
 page:=1
 token:=0
 var prev,next *widget.Button
 show:=func(obj fyne.CanvasObject){pane.Objects=[]fyne.CanvasObject{obj};pane.Refresh()}
 render:=func(){
  token++;request:=token;requestedPage:=page
  status.SetText("Rendering…")
  show(widget.NewProgressBarInfinite())
  if prev!=nil{prev.Disable();next.Disable()}
  go func(){
   var data []byte
   err:=func()error{
    ctx,cancel:=context.WithTimeout(context.Background(),45*time.Second);defer cancel()
    tmp,err:=os.MkdirTemp("","just-wizard-preview-*");if err!=nil{return err};defer os.RemoveAll(tmp)
    png:=filepath.Join(tmp,"preview.png")
    if html{
     browser:=browserBinary();if browser==""{return fmt.Errorf("Install Chromium or Google Chrome to enable HTML previews")}
     // Load local HTML only; an isolated browser profile is removed afterward.
     uri:=(&url.URL{Scheme:"file",Path:filepath.ToSlash(path)}).String()
     cmd:=exec.CommandContext(ctx,browser,"--headless","--disable-gpu","--disable-background-networking","--no-first-run","--user-data-dir="+filepath.Join(tmp,"profile"),"--window-size=1150,1500","--screenshot="+png,uri)
     if out,err:=cmd.CombinedOutput();err!=nil{return fmt.Errorf("HTML render: %w (%s)",err,strings.TrimSpace(string(out)))}
    }else{
     bin,err:=exec.LookPath("pdftoppm");if err!=nil{return fmt.Errorf("Install Poppler (pdftoppm) to enable PDF previews")}
     n:=fmt.Sprint(requestedPage)
     cmd:=exec.CommandContext(ctx,bin,"-f",n,"-l",n,"-scale-to","1600","-singlefile","-png",path,filepath.Join(tmp,"preview"))
     if out,err:=cmd.CombinedOutput();err!=nil{return fmt.Errorf("PDF page %d: %w (%s)",requestedPage,err,strings.TrimSpace(string(out)))}
    }
    data,err=os.ReadFile(png);return err
   }()
   fyne.Do(func(){
    if request!=token{return}
    if err!=nil{
     message:=widget.NewLabel(err.Error());message.Wrapping=fyne.TextWrapWord
     show(container.NewCenter(container.NewVBox(message,widget.NewButton("Open externally",open))))
     status.SetText("Preview unavailable")
    }else{
     image:=canvas.NewImageFromResource(fyne.NewStaticResource("preview.png",data));image.FillMode=canvas.ImageFillContain
     image.SetMinSize(fyne.NewSize(480,650))
     show(container.NewScroll(image))
     if html{status.SetText("HTML viewport preview · open externally for full interactive document")}else{status.SetText(fmt.Sprintf("Page %d",requestedPage))}
    }
    if prev!=nil{if page>1{prev.Enable()};next.Enable()}
   })
  }()
 }
 buttons:=container.NewHBox(widget.NewButton("Open externally",open),widget.NewButton("Refresh",render))
 if !html{
  prev=widget.NewButton("Previous",func(){if page>1{page--;render()}})
  next=widget.NewButton("Next",func(){page++;render()})
  buttons.Add(prev);buttons.Add(next)
 }
 render()
 return container.NewBorder(buttons,status,nil,nil,pane)
}
