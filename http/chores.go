package api

import (
	"crypto/sha1"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"strings"
)

//************************************************* Helpers ************************************************************

func (w welcomeT) hashMessages() string {
	var h hash.Hash
	var toHash string

	for i := range w.Msg {
		toHash += w.Msg[i]
	}
	h = sha1.New()
	io.WriteString(h, toHash)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func getMode(in string) modeT {
	var out modeT

	switch in {
	case "clientSP":
		out = clientSP
	case "clientComp":
		out = clientComp
	case "clientIOT":
		out = clientIOT
	case "ctrlLite":
		out = ctrlLite
	case "ctrlPro":
		out = ctrlPro
	}
	return out
}

func getHostIP(in uint16) string {
	// TODO: Change to compare Client IP with Controller IP
	var out string
	/*
		if hostIP, ok := activeGames[in]; ok {
			out = hostIP
		} else {
			out = "N/A"
		}
	*/
	return out
}

// GetVer returns current (major) version of server
func GetVer() string {
	return version
}

// CloseApp exits the app gracefully
func CloseApp(msg string) {
	fmt.Println(msg)
	os.Exit(0)
}

//************************************************* Safety Measures ****************************************************

func qualifyGET(w http.ResponseWriter, method string, str string) bool {
	if method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		fmt.Println("ERROR:not GET method")
		return false
	}
	return qualifyQuery(w, str)
}

func qualifyPOST(w http.ResponseWriter, method string) bool {
	if method != "POST" {
		http.Error(w, http.StatusText(405), 405)
		fmt.Println("ERROR:not POST method")
		return false
	}
	return true
}

func qualifyPUT(w http.ResponseWriter, method string, str string) bool {
	if method != "PUT" {
		http.Error(w, http.StatusText(405), 405)
		fmt.Println("ERROR:not POST method")
		return false
	}
	return qualifyQuery(w, str)
}

func qualifyQuery(w http.ResponseWriter, in string) bool {
	// prevent empty query string
	if in == "" {
		http.Error(w, http.StatusText(400), 400)
		fmt.Println("ERROR:empty query")
		return false
	}
	// set max chars limit
	if len(in) > 255 {
		http.Error(w, http.StatusText(400), 400)
		fmt.Println("ERROR:query too long")
		return false
	}
	// prevent SQL injection by disallowing apostrophe in query string
	if strings.Index(in, "'") != -1 {
		http.Error(w, http.StatusText(400), 400)
		fmt.Println("ERROR:Message contain illegal character!")
		return false
	}
	// prevent file-structure traversing
	if strings.Index(in, "/") != -1 {
		http.Error(w, http.StatusText(400), 400)
		fmt.Println("ERROR:Message contain illegal character!")
		return false
	}
	return true
}
