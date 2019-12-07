/***********************************************************************************************************************

  qlight.go
  License: MIT
  Copyright (c) 2019 Roy Dybing

  github	: rDybing
  Linked In	: Roy Dybing
  Twitter	: @DybingRoy
  MeWe		: Roy Dybing

  Full license text in README.md

***********************************************************************************************************************/

package main

import (
	"fmt"
	"time"

	api "github.com/rDybing/qlightAPI/http"
)

func main() {
	var input string
	quit := false

	go api.InitAPI()

	time.Sleep(time.Millisecond * 1000)
	help()

	for !quit {
		fmt.Scanf("%s\n", &input)
		switch input {
		case "help":
			help()
		case "welcome":
			api.UpdateWelcome()
		case "quit":
			quit = true
		}
	}
	api.CloseApp("Bahbah")
}

func help() {
	ver := api.GetVer()
	fmt.Println("------------------------")
	fmt.Printf("--  qlightAPI %s  --\n", ver)
	fmt.Println("------------------------")
	fmt.Println("Available Commands:")
	fmt.Println(" - help        | list of commands")
	fmt.Println(" - welcome     | reloads welcomeFile")
	fmt.Println(" - quit        | exit this application")
}

func getInput(in string) string {
	var input string
	fmt.Println(in)
	fmt.Scanf("%s\n", &input)
	return input
}
