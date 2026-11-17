package main

import "strings"

// Canonical icon IDs are shared by migration and desktop rendering.
func validIcon(name string) bool {
 switch name {
 case "terminal", "package", "hammer", "flask-conical", "clipboard-check", "activity", "trash-2", "download", "upload", "tag", "monitor", "list", "settings", "square", "refresh-cw", "archive", "play": return true
 }
 return false
}

func commandIcon(name string) string {
 name = strings.ToLower(strings.TrimSpace(name))
 if i := strings.LastIndex(name, "::"); i >= 0 { name = name[i+2:] }
 name = strings.ReplaceAll(name, "_", "-")
 matches := func(words ...string) bool {
  for _, word := range words {
   if name == word || strings.HasPrefix(name, word+"-") { return true }
  }
  return false
 }
 switch {
 case matches("wizard-terminal", "terminal", "shell", "console", "cli"): return "terminal"
 case matches("bootstrap", "setup", "init", "deps", "requirements"): return "package"
 case matches("build", "compile", "make"): return "hammer"
 case matches("test", "spec", "pytest"): return "flask-conical"
 case matches("ci", "check", "verify", "validate", "lint", "audit"): return "clipboard-check"
 case matches("doctor", "health", "status"): return "activity"
 case matches("clean", "delete", "remove", "prune", "uninstall"): return "trash-2"
 case matches("install", "download", "pull", "fetch"): return "download"
 case matches("deploy", "publish", "release", "push", "upload"): return "upload"
 case matches("version", "tag"): return "tag"
 case matches("wizard", "gui", "desktop", "preview", "show", "serve"): return "monitor"
 case matches("default", "help", "list"): return "list"
 case matches("config", "configure", "settings", "set"): return "settings"
 case matches("stop", "cancel"): return "square"
 case matches("restart", "refresh", "update"): return "refresh-cw"
 case matches("package", "export", "backup"): return "archive"
 case matches("run","start","execute"): return "play"
 case matches("migrate","repair","sync"): return "refresh-cw"
 }
 return ""
}

func groupIconID(group string) string {
 // The last segment is the actual group for module-qualified recipes.
 parts := strings.Split(group, "/")
 group = strings.ToLower(strings.TrimSpace(parts[len(parts)-1]))
 switch group {
 case "development", "dev", "quality", "testing", "tests", "qa", "validation": return "clipboard-check"
 case "build", "compilation": return "hammer"
 case "setup", "installation", "dependencies": return "package"
 case "maintenance", "administration", "admin", "configuration", "settings": return "settings"
 case "application", "applications", "gui", "desktop": return "monitor"
 case "publishing", "deployment", "release", "distribution": return "upload"
 case "information", "documentation", "help": return "list"
 case "books", "content", "assets", "library", "packaging", "backup": return "archive"
 }
 return commandIcon(group)
}

// Explicit valid metadata wins, then the command, then its group.
func (r recipe) iconID() string {
 if id := strings.ToLower(strings.TrimSpace(r.Icon)); validIcon(id) { return id }
 if id := commandIcon(r.Name); id != "" { return id }
 if id := groupIconID(r.Group); id != "" { return id }
 return "play"
}
