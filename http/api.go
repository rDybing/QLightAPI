package api

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/goji/httpauth"
	"github.com/gorilla/mux"
)

const version = "v0.3.1"

var appList appListT
var welcome welcomeT

//************************************************* Server Startup *****************************************************

// InitAPI sets up the endpoints and spins up the API server. Assuming TLS certificates are available, API will use
// https on port 433 only (tlsOK is true). Should TLS certificates not be available, or config file sets server to
// local, server will use http on port 80 (public) or 8080 (local).
func InitAPI() {
	var config configT

	appList = make(map[string]appInfoT)
	appList.loadAppList()

	fmt.Printf("App Entries in AppList: %d\n", len(appList))

	tlsOK := config.loadConfig()
	config.getServerIP(tlsOK)

	if !welcome.loadWelcome() {
		fmt.Println("No welcome messages loaded")
	}

	router := mux.NewRouter()

	server := &http.Server{
		Handler:      router,
		Addr:         config.serverIP,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	router.Handle("/postAppInfo", httpauth.SimpleBasicAuth(config.AuthID, config.AuthKey)(http.HandlerFunc(appList.postAppInfo)))
	router.Handle("/postAppUpdate", httpauth.SimpleBasicAuth(config.AuthID, config.AuthKey)(http.HandlerFunc(appList.postAppUpdate)))
	router.Handle("/getControllerIP", httpauth.SimpleBasicAuth(config.AuthID, config.AuthKey)(http.HandlerFunc(appList.getControllerIP)))
	router.Handle("/getWelcome", httpauth.SimpleBasicAuth(config.AuthID, config.AuthKey)(http.HandlerFunc(welcome.getWelcome)))

	if tlsOK {
		fmt.Println("TLS Certs loaded - running over https")
		fmt.Printf("API    : %s\n", config.serverIP)
		log.Fatal(server.ListenAndServeTLS(config.FullChain, config.PrivKey))
	} else {
		fmt.Println("No TLS Certs - running over http")
		fmt.Printf("API    : %s\n", config.serverIP)
		log.Fatal(server.ListenAndServe())
	}
}

func (c *configT) getServerIP(tlsOK bool) {
	var ip string
	if !c.Local {
		conn, err := net.Dial("udp", "8.8.8.8:80")
		defer conn.Close()
		if err != nil {
			log.Printf("No internet, local only: %v\n", err)
			ip = "127.0.0.1:8080"
		} else {
			port := "80"
			if tlsOK {
				port = "443"
			}
			localIP := conn.LocalAddr().(*net.UDPAddr)
			ip = fmt.Sprintf("%v:%s", localIP.IP, port)
		}
	} else {
		ip = "127.0.0.1:8080"
	}
	c.serverIP = ip
}

//************************************************* Post Calls *********************************************************

func (al appListT) postAppInfo(w http.ResponseWriter, r *http.Request) {
	loc := "postAppInfo"
	fmt.Printf("package: api			func: %s", loc)

	var aTemp appInfoT
	var modeTemp string

	t := time.Now()
	method := r.Method

	if qualifyPOST(w, method) {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ERROR:Wrong data format, %v", err)
		} else {
			aTemp.ID = r.FormValue("ID")
			aTemp.Name = r.FormValue("Name")
			aTemp.WH = r.FormValue("WH")
			aTemp.Aspect = r.FormValue("Aspect")
			aTemp.OS = r.FormValue("OS")
			aTemp.Model = r.FormValue("Model")
			modeTemp = r.FormValue("Mode")
			update := false
			if _, found := al[aTemp.ID]; found {
				aTemp = al[aTemp.ID]
				aTemp.LastLogin = t.UTC().Format("2006-01-02 15:04:05")
				aTemp.LastUpdate = aTemp.LastLogin
				aTemp.Logins++
				aTemp.LastMode = getMode(modeTemp)
				update = true
			} else {
				aTemp.FirstLogin = t.UTC().Format("2006-01-02 15:04:05")
				aTemp.Logins = 1
				aTemp.LastLogin = aTemp.FirstLogin
				aTemp.LastUpdate = aTemp.LastLogin
				aTemp.LastMode = getMode(modeTemp)
			}
			publicIP := r.RemoteAddr
			pubIP := strings.Split(publicIP, ":")
			aTemp.LastPublicIP = pubIP[0]
			aTemp.LastPrivateIP = r.FormValue("PrivateIP")
			al.transferAndSave(aTemp)
			msg := "New Entry"
			if update {
				msg = "Updated Entry"
			}
			fmt.Fprintf(w, "OK:%s", msg)
		}
	}
	fmt.Printf("\tAppID: %s\tAppIP: %s\n", aTemp.ID, aTemp.LastPublicIP)
}

func (al appListT) postAppUpdate(w http.ResponseWriter, r *http.Request) {
	loc := "postUpdateApp"
	fmt.Printf("package: api			func: %s\n", loc)

	var l loggerT
	var aTemp appInfoT
	var out string
	var log bool

	t := time.Now()
	method := r.Method

	if qualifyPOST(w, method) {
		if err := r.ParseForm(); err != nil {
			log = true
			out = "ERROR:Wrongdata format"
		} else {
			idTemp := r.FormValue("ID")
			nameTemp := r.FormValue("Name")
			modeTemp := r.FormValue("Mode")
			if _, found := al[idTemp]; found {
				aTemp = al[idTemp]
				aTemp.LastUpdate = t.UTC().Format("2006-01-02 15:04:05")
				aTemp.Name = nameTemp
				aTemp.LastMode = getMode(modeTemp)
				al.transferAndSave(aTemp)
				out = "OK:Updated Entry"
			} else {
				log = true
				out = "ERROR:No Entry Found"
			}
		}
	}
	if log {
		l.Function = loc
		l.Status = out
		go l.logger()
	} else {
		fmt.Fprintf(w, out)
	}
}

//************************************************* Get Calls **********************************************************

func (al appListT) getControllerIP(w http.ResponseWriter, r *http.Request) {
	var l loggerT
	loc := "getServerIP"
	fmt.Printf("package: api			func: %s\n", loc)

	var client ipT

	method := r.Method

	l.AppID = r.FormValue("ID")

	client.public = r.RemoteAddr
	pubIP := strings.Split(client.public, ":")
	client.public = pubIP[0]

	if _, found := al[l.AppID]; found {
		client.private = al[l.AppID].LastPrivateIP
	}

	out := "ERROR:No LAN server found on IP\n" + client.public
	found := false

	if status, ok := qualifyGET(w, method, l.AppID); ok {
		for i := range al {
			if al[i].LastMode == ctrlLite || al[i].LastMode == ctrlPro {
				if al[i].LastPublicIP == client.public {
					out, found = al.compareFirstThreeIPDigits(i, client)
				}
				if found {
					break
				}
			}
		}
	} else {
		l.Function = loc
		l.Status = status
		out = status
		go l.logger()
	}
	fmt.Fprintf(w, out)
}

func (welcome welcomeT) getWelcome(w http.ResponseWriter, r *http.Request) {
	var l loggerT
	loc := "getWelcome"
	fmt.Printf("package: api			func: %s\n", loc)

	var out string
	method := r.Method

	l.AppID = r.FormValue("ID")
	if status, ok := qualifyGET(w, method, l.AppID); ok {
		// english only for now
		rnd := rand.Intn(len(welcome.Msg))
		out = "OK:" + welcome.Msg[rnd]
	} else {
		l.Function = loc
		l.Status = status
		out = status
		go l.logger()
	}
	fmt.Fprintf(w, out)
}
