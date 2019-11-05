package api

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

const configFile = "./config.json"
const appListFile = "./appList.json"
const stateFile = "./state.json"
const welcomeFile = "./welcome.json"

//************************************************* ConfigFile *********************************************************

func (c *configT) loadConfig() bool {
	var msg string
	ok := true
	f, err := os.Open(configFile)
	if err != nil {
		msg = "Failed to load config file - no https for you!"
		ok = false
	}
	defer f.Close()

	configJSON := json.NewDecoder(f)
	if err = configJSON.Decode(&c); err != nil {
		msg = "Failed to decode config JSON - no https for you!"
		ok = false
	}
	// get if tls certs exist on server
	if _, err := os.Stat(c.FullChain); err != nil {
		msg = "Failed to find FullChain cert - no https for you!"
		ok = false
	}
	if _, err := os.Stat(c.PrivKey); err != nil {
		msg = "Failed to find Private Key - no https for you!"
		ok = false
	}
	if !ok {
		t := time.Now()
		log.WithFields(log.Fields{
			"date":     t.UTC().Format("2006-01-02 15:04:05"),
			"package":  "api",
			"file":     "file.go",
			"function": "loadConfig",
			"error":    err,
			"data":     msg,
		}).Warning(msg)
	}
	return ok
}

//************************************************* Startup Messages Files *********************************************

func (w *welcomeT) loadWelcome() bool {
	var msg string
	ok := true
	f, err := os.Open(welcomeFile)
	if err != nil {
		msg = "Failed to load welcome messages file"
		ok = false
	}
	defer f.Close()

	stateJSON := json.NewDecoder(f)
	if err = stateJSON.Decode(&w); err != nil {
		msg = "Failed to decode welcome messages JSON"
		ok = false
	}
	if !ok {
		t := time.Now()
		log.WithFields(log.Fields{
			"date":     t.UTC().Format("2006-01-02 15:04:05"),
			"package":  "api",
			"file":     "file.go",
			"function": "loadWelcome",
			"error":    err,
			"data":     msg,
		}).Warning(msg)
	}
	return ok
}

//************************************************* AppList ************************************************************

func (al appListT) loadAppList() {
	var appInfo []appInfoT

	f, err := os.Open(appListFile)
	if err != nil {
		log.Println("Failed to load appList file")
	}
	defer f.Close()

	jsonRaw := json.NewDecoder(f)
	if err := jsonRaw.Decode(&appInfo); err != nil {
		t := time.Now()
		log.WithFields(log.Fields{
			"date":     t.UTC().Format("2006-01-02 15:04:05"),
			"package":  "api",
			"file":     "file.go",
			"function": "loadApplist",
			"error":    err,
			"data":     jsonRaw,
		}).Warning("Failed to decode appList JSON")
	}
	for i := range appInfo {
		wg.Add(1)
		go al.transfer(appInfo[i])
		wg.Wait()
	}
}

func (al appListT) saveAppList() {
	var temp appInfoT
	var out []appInfoT

	for i := range al {
		temp = al[i]
		out = append(out, temp)
	}
	outBytes, err := json.MarshalIndent(out, "", "	")
	if err != nil {
		t := time.Now()
		log.WithFields(log.Fields{
			"date":     t.UTC().Format("2006-01-02 15:04:05"),
			"package":  "api",
			"file":     "file.go",
			"function": "saveApplist",
			"error":    err,
			"data":     outBytes,
		}).Error("Failed to encode appList map")
	}
	outJSON := string(outBytes[:])
	f, err := os.OpenFile(appListFile, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()
	if err != nil {
		t := time.Now()
		log.WithFields(log.Fields{
			"date":     t.UTC().Format("2006-01-02 15:04:05"),
			"package":  "api",
			"file":     "file.go",
			"function": "saveApplist",
			"error":    err,
			"data":     outJSON,
		}).Error("Failed to save appList")
	} else {
		if outJSON != "null" {
			fmt.Fprintf(f, outJSON)
		} else {
			fmt.Println("No data in AppList to save")
		}
	}
}
