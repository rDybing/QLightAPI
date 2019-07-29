package api

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const appListFile = "appList.json"

//************************************************* AppsList ***********************************************************
/*
// LoadAppList Loads up file of state of aux-apps
func LoadAppList() ([]AppListT, error) {
	var AppList []AppListT

	f, err := os.OpenFile(appListFile, os.O_RDONLY, 0666)
	if err != nil {
		log.Printf("ERROR:No auxAppList file exist, %v\n", err)
		return nil, err
	}
	defer f.Close()

	jsonRaw := json.NewDecoder(f)
	if err := jsonRaw.Decode(&AppList); err != nil {
		if err == io.EOF {
			return nil, err
		}
		log.Printf("ERROR: Parsing AppList JSON, %v\n", err)
		return nil, err
	}
	return AppList, nil
}
*/

// SaveAppList Saves the log file of apps/devices that have used this API
// Call at new entry and at exit
func (a AppListT) SaveAppList() {
	var temp AppListT
	var out []AppListT

	for i := range in {
		temp = in[i]
		out = append(out, temp)
	}
	outBytes, err := json.MarshalIndent(out, "", "	")
	if err != nil {
		log.Printf("Could not JSONify AppList, %v", err)
	}
	outJSON := string(outBytes[:])
	f, err := os.OpenFile(appListFile, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("ERROR:Log File is being very stubborn, %v\n", err)
	} else {
		if outJSON != "null" {
			fmt.Fprintf(f, outJSON)
			fmt.Println(outJSON)
		} else {
			fmt.Println("No logging data in AppList to save")
		}
	}
	defer f.Close()
}
