package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var wg sync.WaitGroup
var guard sync.Mutex

//************************************************* Public Helpers *****************************************************

// UpdateWelcome reloads the welcomeFile in order to renew without restarting the API-app
func UpdateWelcome() {
	fmt.Println("Reloading welcome messages")
	welcome.loadWelcome()
	fmt.Printf("Messages loaded: %d\n", len(welcome.Msg))
}

// GetVer returns current (major) version of API-app
func GetVer() string {
	return version
}

// CloseApp exits the API-app gracefully and saves current in-memory appList
func CloseApp(msg string) {
	fmt.Println("...saving appList")
	appList.saveAppList()
	fmt.Println(msg)
	os.Exit(0)
}

//************************************************* Private Helpers ****************************************************

func getMode(in string) modeT {
	var out modeT

	switch in {
	case "clientSP":
		// 0: Client mode SmartPhone
		out = clientSP
	case "clientComp":
		// 1: Client mode Computer
		out = clientComp
	case "clientIOT":
		// 2: Client mode Arduino
		out = clientIOT
	case "ctrlLite":
		// 3: Controller mode Free version
		out = ctrlLite
	case "ctrlPro":
		// 4: Controller mode Pro version
		out = ctrlPro
	case "noneSP":
		// 5: Mode not selected SmartPhone
		out = noneSP
	case "noneComp":
		// 6: Mode not selected Computer
		out = noneComp
	}
	return out
}

func (l *loggerT) logger() {
	t := time.Now()
	l.Date = t.UTC().Format("2006-01-02 15:04:05")

	outBytes, err := json.MarshalIndent(l, "", "	")
	if err != nil {
		log.WithFields(log.Fields{
			"date":     l.Date,
			"package":  "api",
			"file":     "chores.go",
			"function": "logger",
			"error":    err,
			"data":     outBytes,
		}).Warning("ERROR:Could not JSONify log-entry")
	}
	outJSON := string(outBytes[:])

	fmt.Println(outJSON)
}

func (al appListT) compareFirstThreeIPDigits(index string, c ipT) (string, bool) {
	out := "ERROR:Not Found"
	found := false

	clientPIP := strings.Split(c.private, ".")
	c.private = fmt.Sprintf("%s.%s.%s", clientPIP[0], clientPIP[1], clientPIP[2])
	serverPIPFull := al[index].LastPrivateIP
	serverPIP := strings.Split(serverPIPFull, ".")
	serverPrivateIP := fmt.Sprintf("%s.%s.%s", serverPIP[0], serverPIP[1], serverPIP[2])
	if c.private == serverPrivateIP {
		found = true
		out = "OK:" + serverPIPFull
	}
	return out, found
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
	// prevent SQL injection by disallowing semi-colon in query string
	if strings.Index(in, ";") != -1 {
		http.Error(w, http.StatusText(400), 400)
		status := "ERROR:Message contain illegal character!"
		fmt.Println(status)
		return status, false
	}
	// prevent file-structure traversing
	if strings.Index(in, "/") != -1 || strings.Index(in, "\\") != -1 {
		http.Error(w, http.StatusText(400), 400)
		status := "ERROR:Message contain illegal character!"
		fmt.Println()
		return status, false
	}
	return "", true
}
