package main

import (
 "os"
 "path/filepath"
 "testing"
)

func TestExplorerSaveValidationAndConflict(t *testing.T) {
 path:=filepath.Join(t.TempDir(),"data.json")
 original:=[]byte("{\"ready\":false}\n")
 if err:=os.WriteFile(path,original,0600);err!=nil{t.Fatal(err)}
 if err:=saveExplorerText(path,original,[]byte("{invalid"));err==nil{t.Fatal("invalid JSON accepted")}
 updated:=[]byte("{\"ready\":true}\n")
 if err:=saveExplorerText(path,original,updated);err!=nil{t.Fatal(err)}
 if err:=saveExplorerText(path,original,original);err==nil{t.Fatal("external change overwritten")}
 data,err:=os.ReadFile(path);if err!=nil||string(data)!=string(updated){t.Fatalf("unexpected saved data: %q %v",data,err)}
}
func TestExplorerYamlValidation(t *testing.T){
 if err:=validateExplorerText("config.yaml",[]byte("layout: advent\n"));err!=nil{t.Fatal(err)}
 if err:=validateExplorerText("config.yaml",[]byte("layout: [\n"));err==nil{t.Fatal("invalid YAML accepted")}
}
