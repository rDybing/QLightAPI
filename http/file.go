package api

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const configFile = "./config.json"
const appListFile = "./appList.json"
const stateFile = "./state.json"
const welcomeFile = "./welcome.json"

//************************************************* ConfigFile *********************************************************

func (c *configT) loadConfig() bool {
	ok := true
	f, err := os.Open(configFile)
	if err != nil {
		log.Printf("Failed to load config file - no https for you!\n%v\n", err)
		ok = false
	}
	defer f.Close()

	configJSON := json.NewDecoder(f)
	if err = configJSON.Decode(&c); err != nil {
		log.Printf("Failed to decode config JSON - no https for you!\n%v\n", err)
		ok = false
	}
	// get if tls certs exist on server
	if _, err := os.Stat(c.FullChain); err != nil {
		log.Printf("Failed to find FullChain cert - no https for you!\n%v\n", err)
		ok = false
	}
	if _, err := os.Stat(c.PrivKey); err != nil {
		log.Printf("Failed to find Private Key - no https for you!\n%v\n", err)
		ok = false
	}
	return ok
}

//************************************************* Startup Messages Files *********************************************

func (s *stateT) loadState() bool {
	ok := true
	f, err := os.Open(stateFile)
	if err != nil {
		log.Printf("Failed to load state file")
		ok = false
	}
	defer f.Close()

	stateJSON := json.NewDecoder(f)
	if err = stateJSON.Decode(&s); err != nil {
		log.Printf("Failed to decode state JSON")
		ok = false
	}
	return ok
}

func (s stateT) saveState(hash string) error {
	outBytes, err := json.MarshalIndent(s, "", "	")
	if err != nil {
		log.Printf("ERROR:Could not JSONify state, %v", err)
	}
	outJSON := string(outBytes[:])
	f, err := os.OpenFile(stateFile, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("ERROR:State File is being very stubborn, %v\n", err)
	} else {
		if outJSON != "null" {
			fmt.Fprintf(f, outJSON)
			//fmt.Println(outJSON)
		} else {
			fmt.Println("No data in state to save")
		}
	}
	defer f.Close()
	return err
}

func (w *welcomeT) loadWelcome() bool {
	ok := true
	f, err := os.Open(welcomeFile)
	if err != nil {
		log.Printf("Failed to load welcome messages file")
		ok = false
	}
	defer f.Close()

	stateJSON := json.NewDecoder(f)
	if err = stateJSON.Decode(&w); err != nil {
		log.Printf("Failed to decode welcome messages JSON")
		ok = false
	}
	return ok
}

func saveWelcome() {
	var out welcomeT

	out.Msg = append(out.Msg, "Herding server hamsters...")
	out.Msg = append(out.Msg, "Swatting new and interesting bugs!")

	outBytes, err := json.MarshalIndent(out.Msg, "", "	")
	if err != nil {
		log.Printf("ERROR:Could not JSONify AppList, %v", err)
	}
	outJSON := string(outBytes[:])
	f, err := os.OpenFile(welcomeFile, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("ERROR:welcome File is being very stubborn, %v\n", err)
	} else {
		if outJSON != "null" {
			fmt.Fprintf(f, outJSON)
			//fmt.Println(outJSON)
		} else {
			fmt.Println("No data in welcome to save")
		}
	}
	defer f.Close()

}

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

func (al appListT) SaveAppList() {
	var temp appInfoT
	var out []appInfoT

	for i := range al {
		temp = al[i]
		out = append(out, temp)
	}
	outBytes, err := json.MarshalIndent(out, "", "	")
	if err != nil {
		log.Printf("ERROR:Could not JSONify AppList, %v", err)
	}
	outJSON := string(outBytes[:])
	f, err := os.OpenFile(appListFile, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("ERROR:Applist File is being very stubborn, %v\n", err)
	} else {
		if outJSON != "null" {
			fmt.Fprintf(f, outJSON)
		} else {
			fmt.Println("No data in AppList to save")
		}
	}
	defer f.Close()
}
