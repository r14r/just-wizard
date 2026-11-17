package main

import (
 "bytes"
 "encoding/json"
 "encoding/xml"
 "fmt"
 "io"
 "os"
 "path/filepath"
 "strings"

 "gopkg.in/yaml.v3"
)

func validateExplorerText(path string, data []byte) error {
 switch strings.ToLower(filepath.Ext(path)){
 case ".json":
  if !json.Valid(data){return fmt.Errorf("Invalid JSON syntax")}
 case ".yaml",".yml":
  d:=yaml.NewDecoder(bytes.NewReader(data))
  for {var node yaml.Node;err:=d.Decode(&node);if err==io.EOF{break};if err!=nil{return err}}
 case ".xml",".svg":
  d:=xml.NewDecoder(bytes.NewReader(data))
  for {_,err:=d.Token();if err==io.EOF{break};if err!=nil{return err}}
 }
 return nil
}

// Save only if the on-disk file still matches the version originally loaded.
func saveExplorerText(path string, original, updated []byte) error {
 if err:=validateExplorerText(path,updated);err!=nil{return err}
 info,err:=os.Lstat(path);if err!=nil{return err}
 if !info.Mode().IsRegular(){return fmt.Errorf("Not a regular file")}
 current,err:=os.ReadFile(path);if err!=nil{return err}
 if !bytes.Equal(current,original){return fmt.Errorf("File changed on disk. Reopen before saving.")}
 f,err:=os.CreateTemp(filepath.Dir(path),".explorer-save-*");if err!=nil{return err}
 defer os.Remove(f.Name())
 defer f.Close()
 if err=f.Chmod(info.Mode().Perm());err!=nil{return err}
 if _,err=f.Write(updated);err!=nil{return err}
 if err=f.Sync();err!=nil{return err}
 if err=f.Close();err!=nil{return err}
 return os.Rename(f.Name(),path)
}
