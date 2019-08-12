package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup
var guard sync.Mutex

//************************************************* Helpers ************************************************************

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
	case "noneSP":
		out = noneSP
	case "noneComp":
		out = noneComp
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
	fmt.Println("...saving appList")
	appList.saveAppList()
	fmt.Println(msg)
	os.Exit(0)
}

func (l *loggerT) logger() {
	t := time.Now()
	l.Date = t.UTC().Format("2006-01-02 15:04:05")

	outBytes, err := json.MarshalIndent(l, "", "	")
	if err != nil {
		log.Printf("ERROR:Could not JSONify AppList, %v", err)
	}
	outJSON := string(outBytes[:])

	fmt.Println(outJSON)
}

func (al appListT) transferAndSave(ai appInfoT) {
	wg.Add(1)
	go al.transfer(ai)
	wg.Wait()
	guard.Lock()
	go al.saveAppList()
	guard.Unlock()
}

func (al appListT) transfer(ai appInfoT) {
	al[ai.ID] = ai
	wg.Done()
}

//************************************************* Safety Measures ****************************************************

func qualifyGET(w http.ResponseWriter, method string, str string) (string, bool) {
	if method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		status := "ERROR:not GET method"
		fmt.Println(status)
		return status, false
	}
	status, ok := qualifyQuery(w, str)
	return status, ok
}

func qualifyPOST(w http.ResponseWriter, method string) bool {
	if method != "POST" {
		http.Error(w, http.StatusText(405), 405)
		fmt.Println("ERROR:not POST method")
		return false
	}
	return true
}

func qualifyPUT(w http.ResponseWriter, method string, str string) (string, bool) {
	if method != "PUT" {
		http.Error(w, http.StatusText(405), 405)
		status := "ERROR:not POST method"
		fmt.Println(status)
		return status, false
	}
	status, ok := qualifyQuery(w, str)
	return status, ok
}

func qualifyQuery(w http.ResponseWriter, in string) (string, bool) {
	// prevent empty query string
	if in == "" {
		http.Error(w, http.StatusText(400), 400)
		status := "ERROR:empty query"
		fmt.Println(status)
		return status, false
	}
	// set max chars limit
	if len(in) > 255 {
		http.Error(w, http.StatusText(400), 400)
		status := "ERROR:query too long"
		fmt.Println(status)
		return status, false
	}
	// prevent SQL injection by disallowing apostrophe in query string
	if strings.Index(in, "'") != -1 {
		http.Error(w, http.StatusText(400), 400)
		status := "ERROR:Message contain illegal character!"
		fmt.Println(status)
		return status, false
	}
	// prevent file-structure traversing
	if strings.Index(in, "/") != -1 {
		http.Error(w, http.StatusText(400), 400)
		status := "ERROR:Message contain illegal character!"
		fmt.Println()
		return status, false
	}
	return "", true
}
