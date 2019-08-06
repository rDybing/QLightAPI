package api

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
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
	fmt.Println("...saving appList")
	appList.SaveAppList()
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

	log.Println(outJSON)
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
