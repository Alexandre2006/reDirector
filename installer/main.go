package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		help()
		return
	}

	switch os.Args[1] {
	case "install":
		install()
	case "uninstall":
		uninstall()
	case "repair":
		repair()
	case "help":
		help()
	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		help()
	}
}

func help() {
	fmt.Println(`reDirector

Usage:
  reDirector <command>

Commands:
  install    installs reDirector
  uninstall  uninstalls reDirector completely
  repair     fixes reDirector after device updates
  help       shows this message`)
}
