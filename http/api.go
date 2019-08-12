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

const version = "v0.2.0"

var appList appListT

//************************************************* Server Startup *****************************************************

// InitAPI sets up the endpoints and spins up the API server. Assuming TLS certificates are available, API will use
// https on port 433 only (tlsOK is true). Should TLS certificates not be available, or config file sets server to
// local, server will use http on port 80 (public) or 8080 (local).
func InitAPI() {
	var config configT
	var welcome welcomeT

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
	router.Handle("/putAppUpdate", httpauth.SimpleBasicAuth(config.AuthID, config.AuthKey)(http.HandlerFunc(appList.putAppUpdate)))
	router.Handle("/getServerIP", httpauth.SimpleBasicAuth(config.AuthID, config.AuthKey)(http.HandlerFunc(appList.getServerIP)))
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
	fmt.Printf("package: api			func: %s\n", loc)

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
			publicIP := r.RemoteAddr
			pubIP := strings.Split(publicIP, ":")
			aTemp.LastPublicIP = pubIP[0]
			aTemp.LastPrivateIP = r.FormValue("PrivateIP")
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
			al.transferAndSave(aTemp)
			msg := "New Entry"
			if update {
				msg = "Updated Entry"
			}
			fmt.Fprintf(w, "OK:%s", msg)
		}
	}
}

//************************************************* Put Calls *********************************************************

func (al appListT) putAppUpdate(w http.ResponseWriter, r *http.Request) {
	loc := "putUpdateApp"
	fmt.Printf("package: api			func: %s\n", loc)

	var l loggerT
	var aTemp appInfoT
	var out string

	t := time.Now()
	method := r.Method

	if err := r.ParseForm(); err != nil {
		l.Function = loc
		l.Status = "ERROR:Wrongdata format"
		out = l.Status
		go l.logger()
	} else {
		idTemp := r.FormValue("ID")
		nameTemp := r.FormValue("Name")
		modeTemp := r.FormValue("Mode")
		checkTemp := idTemp + nameTemp + modeTemp
		if status, ok := qualifyPUT(w, method, checkTemp); ok {
			if _, found := al[aTemp.ID]; found {
				aTemp = al[aTemp.ID]
				aTemp.LastUpdate = t.UTC().Format("2006-01-02 15:04:05")
				aTemp.Name = nameTemp
				aTemp.LastMode = getMode(modeTemp)
			}
			al.transferAndSave(aTemp)
			out = "OK:Updated Entry"
		} else {
			l.Function = loc
			l.Status = status
			out = status
			go l.logger()
		}
	}
	fmt.Fprintf(w, out)
}

//************************************************* Get Calls **********************************************************

func (al appListT) getServerIP(w http.ResponseWriter, r *http.Request) {
	var l loggerT
	loc := "getServerIP"
	fmt.Printf("package: api			func: %s\n", loc)

	method := r.Method

	clientPrivateIP := r.FormValue("privateIP")
	clientPublicIP := r.RemoteAddr

	pubIP := strings.Split(clientPublicIP, ":")
	clientPublicIP = pubIP[0]

	out := "ERROR:No LAN server found on IP\n" + clientPublicIP

	found := false
	if status, ok := qualifyGET(w, method, clientPrivateIP); ok {
		for i := range al {
			if al[i].LastPublicIP == clientPublicIP {
				clientPIP := strings.Split(clientPrivateIP, ".")
				clientPrivateIP = fmt.Sprintf("%s.%s.%s", clientPIP[0], clientPIP[1], clientPIP[2])
				serverPrivateIPFull := al[i].LastPrivateIP
				serverPIP := strings.Split(serverPrivateIPFull, ".")
				serverPrivateIPComp := fmt.Sprintf("%s.%s.%s", serverPIP[0], serverPIP[1], serverPIP[2])
				if clientPrivateIP == serverPrivateIPComp {
					found = true
					out = "OK:" + serverPrivateIPFull
					break
				}
			}
		}
		if !found {
			out = "ERROR:Not Found"
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

	l.AppID = r.FormValue("appID")
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
