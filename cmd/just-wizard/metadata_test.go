package main
import("testing";"strings";"reflect")
func TestUIMetadata(t *testing.T){
 src:=[]byte("# @ui group Books\n# @ui label Build a book\n# @ui description Build one book.\n# @ui param book type=number min=1 max=99 label=Book_ID\n# @ui param layout type=select options=advent|modern|classic default=advent\n# @ui confirm\n[group('Build')]\nbuild-book book layout=\"advent\":\n    echo ok\n")
 rs:=applyMetadata([]recipe{{Name:"build-book",Group:"Build",Params:[]parameter{{Name:"book"},{Name:"layout"}}}},src)
 r:=rs[0];if r.Group!="Books"||r.title()!="Build a book"||!r.Confirm||r.Params[0].label()!="Book ID"{t.Fatalf("%+v",r)}
 a,e:=recipeArgs(r,[]string{"12",""});if e!=nil||!reflect.DeepEqual(a,[]string{"build-book","12","advent"}){t.Fatalf("%v %v",a,e)}
 for _,v:=range [][]string{{"100","advent"},{"1","bad"},{"NaN","modern"}}{if _,e=recipeArgs(r,v);e==nil{t.Fatalf("accepted %v",v)}}
}
func TestMigrateSource(t *testing.T){
 src:=[]byte("# existing comment\r\n[group('Build')]\nbuild target=\"linux\":\r\n    echo hi\r\n")
 rs:=[]recipe{{Name:"build",Group:"Build",Params:[]parameter{{Name:"target"}}}}
 // Use uniform line endings as ordinary Justfiles do.
 src=[]byte(strings.ReplaceAll(strings.ReplaceAll(string(src),"\r\n","\n"),"\n","\r\n"))
 a,n:=migrateSource(src,rs);if n!=1||!strings.Contains(string(a),"# @ui group Build\r\n")||!strings.Contains(string(a),"[group('Build')]\r\nbuild target="){t.Fatalf("%d %s",n,a)}
 b,n:=migrateSource(a,rs);if n!=0||string(a)!=string(b){t.Fatal("migration not idempotent")}
}
