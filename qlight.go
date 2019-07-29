/********************************
  qlight.go
  License: MIT
  Copyright (c) 2019 Roy Dybing
  github   : rDybing
  Linked In: Roy Dybing
  MeWe     : Roy Dybing
  Full license text in README.md
*********************************/

package main

import (
	api "github.com/rDybing/qlightAPI/http"
)

func main() {
	quit := false

	go api.InitAPI()

	for !quit {

	}
}
