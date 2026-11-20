package main

import (
 "flag"
 "fmt"
 "io"
 "os"
)

const uiIcons = "terminal, package, hammer, flask-conical, clipboard-check, activity, trash-2, download, upload, tag, monitor, list, settings, square, refresh-cw, archive, play"

const optionsHelp = `Just Wizard — supported Justfile UI metadata

Annotations (place directly above a recipe):
  # @ui group NAME                Sidebar group, overrides native [group('NAME')]
  # @ui label TEXT                Display name
  # @ui description TEXT          Explanation shown in the UI
  # @ui icon ICON_ID              Recipe icon (see --icons)
  # @ui confirm                   Ask for confirmation before executing
  # @ui param NAME ATTRIBUTE...   Configure a recipe parameter

Parameter attributes:
  type=text|password|number|path|select|checkbox
  label=TEXT                     Display label (use underscores for spaces)
  description=TEXT               Help text (use underscores for spaces)
  default=VALUE                  Initial value
  required                       Value must not be empty
  options=a|b|c                  Choices for select
  min=NUMBER                     Minimum numeric value
  max=NUMBER                     Maximum numeric value

Usage:
  just-wizard options            Show all annotations, attributes and example
  just-wizard options --example  Print only a sample Justfile
  just-wizard options --icons    List available icon IDs
  just-wizard migrate --dry-run  Preview automatic annotations
  just-wizard migrate            Annotate the current Justfile
`

const optionsExample = `# @ui group Development
# @ui icon flask-conical
# @ui label Run tests
# @ui description Run the selected test suite
# @ui param suite type=select options=all|unit|integration default=all label=Test_suite description=Choose_the_test_suite
test suite="all":
    echo "Testing {{suite}}"

# @ui group Build
# @ui icon hammer
# @ui label Build application
# @ui description Build for a selected platform
# @ui param target type=select options=linux|darwin|windows default=linux label=Target_platform
build target="linux":
    echo "Building for {{target}}"

# @ui group Maintenance
# @ui icon trash-2
# @ui label Clean generated files
# @ui description Remove generated build files
# @ui confirm
clean:
    echo "Cleaning build files"
`

func printOptions(w io.Writer, exampleOnly, iconsOnly bool) {
 if exampleOnly {fmt.Fprint(w,optionsExample);return}
 if iconsOnly {fmt.Fprintln(w,"Supported @ui icon IDs:");fmt.Fprintln(w,uiIcons);return}
 fmt.Fprint(w,optionsHelp)
 fmt.Fprintln(w,"Supported @ui icon IDs:")
 fmt.Fprintln(w,uiIcons)
 fmt.Fprintln(w,"\nExample Justfile:")
 fmt.Fprint(w,optionsExample)
}

func runOptions(args []string) error {
 fs:=flag.NewFlagSet("options",flag.ContinueOnError)
 fs.SetOutput(os.Stderr)
 example:=fs.Bool("example",false,"print a sample Justfile")
 icons:=fs.Bool("icons",false,"list supported icon IDs")
 if err:=fs.Parse(args);err!=nil {if err==flag.ErrHelp{return nil};return err}
 if fs.NArg()>0{return fmt.Errorf("unexpected argument: %s",fs.Arg(0))}
 if *example&&*icons{return fmt.Errorf("--example and --icons are mutually exclusive")}
 printOptions(os.Stdout,*example,*icons)
 return nil
}
