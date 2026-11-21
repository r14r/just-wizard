package main
import("bufio";"flag";"fmt";"os";"os/exec";"strconv";"strings")
func terminalMain(){
 if dispatchMigration(){return}
 root:=flag.String("root",".","project directory");file:=flag.String("justfile","","explicit justfile");flag.Parse()
 dir,jf,rs,err:=loadProject(*root,*file);if err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(1)}
 in:=bufio.NewReader(os.Stdin)
 ask:=func(s string)string{fmt.Print(s);v,e:=in.ReadString('\n');if e!=nil{os.Exit(0)};return strings.TrimSuffix(strings.TrimSuffix(v,"\n"),"\r")}
 choose:=func(title string,items []string)int{fmt.Println("\n"+title);for i,s:=range items{fmt.Printf("%d  %s\n",i+1,s)};n,_:=strconv.Atoi(ask("0 Back / Exit: "));return n-1}
 groups:=recipeGroups(rs)
 for{g:=choose("Just Wizard",groups);if g<0{return};if g>=len(groups){continue};var choices []recipe;var labels []string;for _,r:=range rs{if r.Group==groups[g]{choices=append(choices,r);labels=append(labels,r.title()+" — "+r.Description)}};i:=choose(groups[g],labels);if i<0||i>=len(choices){continue};r:=choices[i];r,e:=resolveParameterSources(r,dir,jf,rs);if e!=nil{fmt.Println("Could not load parameter choices:",e);continue};var vals []string;cancelled:=false;for _,p:=range r.Params{v:="";if len(p.UI.Options)>0{items:=append([]string(nil),p.UI.Options...);offset:=0;if p.optional(){items=append([]string{"Use default / omit"},items...);offset=1};n:=choose(p.label(),items);if n<0||n>=len(items){cancelled=true;break};if n>=offset{v=items[n]}}else{v=ask(p.label()+" ("+p.hint()+"): ")};if v==""{v=p.initial()};vals=append(vals,v)};if cancelled{continue};args,e:=recipeArgs(r,vals);if e!=nil{fmt.Println(e);continue};fmt.Println(safePreview(r,args));if ask("Run recipe? [y/N] ")!="y"{continue};cmd:=exec.Command("just",append([]string{"--justfile",jf,"--working-directory",dir,"--"},args...)...);cmd.Dir=dir;cmd.Stdout=os.Stdout;cmd.Stderr=os.Stderr;cmd.Stdin=os.Stdin;if e=cmd.Run();e!=nil{fmt.Println(e)}}
}
