package api

import (
	"crypto/sha1"
	"fmt"
	"hash"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/goji/httpauth"
	"github.com/gorilla/mux"
)

const version = "v0.1.0"

type modeT int

const (
	clientSP modeT = iota
	clientComp
	clientIOT
	ctrlLite
	ctrlPro
)

type configT struct {
	FullChain string
	PrivKey   string
	Local     bool
	AuthID    string
	AuthKey   string
	serverIP  string
}

type stateT struct {
	welcomeHash string
}

type welcomeT struct {
	Msg []string
}

type appInfoT struct {
	ID            string
	Name          string
	WH            string
	Aspect        string
	LastPublicIP  string
	LastPrivateIP string
	OS            string
	Model         string
	Logins        int
	FirstLogin    string
	LastLogin     string
	LastMode      modeT
}

type appListT map[string]appInfoT

// InitAPI sets up the endpoints and spins up the API server
func InitAPI() {
	var config configT
	var state stateT
	var appList appListT
	var welcome welcomeT

	appList = make(map[string]appInfoT)

	tlsOK := config.loadConfig()
	config.getServerIP(tlsOK)

	if !welcome.loadWelcome() {
		fmt.Println("No welcome messages loaded")
	}

	state.welcomeHash = welcome.hashMessages()
	fmt.Printf("welcome Hash: %s\n", state.welcomeHash)

	router := mux.NewRouter()

	server := &http.Server{
		Handler:      router,
		Addr:         config.serverIP,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	router.Handle("/postAppInfo", httpauth.SimpleBasicAuth(config.AuthID, config.AuthKey)(http.HandlerFunc(appList.postAppInfo)))
	router.Handle("/postControllerIP", httpauth.SimpleBasicAuth(config.AuthID, config.AuthKey)(http.HandlerFunc(postControllerIP)))

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

	var h hash.Hash
	var hash string
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
			aTemp.LastPublicIP = r.RemoteAddr
			aTemp.LastPrivateIP = r.FormValue("PrivateIP")
			aTemp.OS = r.FormValue("OS")
			aTemp.Model = r.FormValue("Model")
			modeTemp = r.FormValue("Mode")

			h = sha1.New()
			io.WriteString(h, aTemp.ID+aTemp.Name)
			hash = fmt.Sprintf("%x", h.Sum(nil))

			if _, found := al[hash]; found {
				aTemp = al[hash]
				aTemp.LastLogin = t.UTC().Format("2006-01-02 15:04:05")
				aTemp.Logins++
				aTemp.LastMode = getMode(modeTemp)
				al[hash] = aTemp
				status := "OK"
				fmt.Fprintf(w, status)
			} else {
				aTemp.FirstLogin = t.UTC().Format("2006-01-02 15:04:05")
				aTemp.Logins = 1
				aTemp.LastLogin = aTemp.FirstLogin
				aTemp.LastMode = getMode(modeTemp)
				al[hash] = aTemp
				// append new entry to file
				go al.SaveAppList()
				// return new registration Created OK
				fmt.Fprintf(w, "NEW")
			}
		}
	}
}

func postControllerIP(w http.ResponseWriter, r *http.Request) {
	loc := "getLocalIP"
	fmt.Printf("package: api			func: %s\n", loc)

	var out string
	method := r.Method

	hostID := r.FormValue("hostID")
	if qualifyGET(w, method, hostID) {
		hostInt, _ := strconv.Atoi(hostID)
		hostUint := uint16(hostInt)
		out = getHostIP(hostUint)
	} else {
		out = "ERROR:Could not parse host ID"
	}
	fmt.Fprintf(w, out)
}
