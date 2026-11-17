package main
import("testing";"reflect")
func TestDiscovery(t *testing.T){
 rs,e:=decodeRecipes([]byte(`{"recipes":{"test":{"doc":"Run tests","attributes":[{"group":"Quality"}],"parameters":[]},"build":{"attributes":[{"group":"Build"}],"parameters":[{"name":"target","kind":"singular","default":null},{"name":"layout","kind":"singular","default":"modern"}]},"hidden":{"private":true}},"modules":{"tools":{"recipes":{"check":{"parameters":[]}}}},"aliases":{"verify":{"target":"test"}}}`))
 if e!=nil{t.Fatal(e)};if len(rs)!=4{t.Fatalf("%+v",rs)}
 if !reflect.DeepEqual(recipeGroups(rs),[]string{"Build","Quality","tools / Ungrouped"}){t.Fatal(recipeGroups(rs))}
 args,e:=recipeArgs(rs[0],[]string{"a b",""});if e!=nil||!reflect.DeepEqual(args,[]string{"build","a b"}){t.Fatalf("%v %v",args,e)}
 if _,e=recipeArgs(rs[0],[]string{"",""});e==nil{t.Fatal("required parameter accepted")}
}
