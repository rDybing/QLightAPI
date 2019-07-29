package api

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/goji/httpauth"
	"github.com/gorilla/mux"
)

const configFile = "./config.json"
const version = "v0.1.0"

type configT struct {
	FullChain string
	PrivKey   string
	Local     bool
	AuthName  string
	AuthKey   string
}

type outPutT struct {
	message  string
	clientIP string
	serverIP string
}

type AppInfoT struct {
	ID         string
	Name       string
	WH         string
	Aspect     string
	LastIP     string
	OS         string
	Model      string
	Logins     int
	FirstLogin string
	LastLogin  string
}

type AppListT struct {
	List map[string]AppInfoT
}

// InitAPI sets up the endpoints and spins up the API server
func InitAPI() {
	var app AppInfoT
	var con configT
	var out outPutT
	var apps AppListT

	apps.List = make(map[string]appInfoT)

	tlsOK := con.loadConfig()
	out.getPrivateIP(con.Local, tlsOK)

	router := mux.NewRouter()

	server := &http.Server{
		Handler:      router,
		Addr:         out.serverIP,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	router.Handle("/post/appInfo/", httpauth.SimpleBasicAuth(con.AuthName, con.AuthKey)(http.HandlerFunc(apps.postAppInfo)))
	router.Handle("/post/controllerIP/", httpauth.SimpleBasicAuth(con.AuthName, con.AuthKey)(http.HandlerFunc(postControllerIP)))

	if tlsOK {
		fmt.Println("TLS Certs loaded - running over https")
		fmt.Printf("API    : %s\n", out.serverIP)
		log.Fatal(server.ListenAndServeTLS(con.FullChain, con.PrivKey))
	} else {
		fmt.Println("No TLS Certs - running over http")
		fmt.Printf("API    : %s\n", out.serverIP)
		log.Fatal(server.ListenAndServe())
	}
}

func (o *outPutT) testPage(w http.ResponseWriter, r *http.Request) {
	ip, port, _ := net.SplitHostPort(r.RemoteAddr)
	o.clientIP = fmt.Sprintf("%s:%s", ip, port)
	fmt.Fprintf(w, "%s", o.message)
	fmt.Printf("Inbound from     : %s\n", o.clientIP)
	fmt.Printf("Response from    : %s\n", o.serverIP)
	fmt.Fprintf(w, "Inbound from     : %s\nResponse from    : %s", o.clientIP, o.serverIP)
}

func (o *outPutT) getPrivateIP(local, tlsOK bool) {
	var ip string
	if !local {
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
	o.serverIP = ip
}

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

//************************************************* Post Calls *********************************************************

func (a *AppListT) postAppInfo(w http.ResponseWriter, r *http.Request) {
	var h hash.Hash
	var hash string
	var aTemp AppInfoT

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
			aTemp.LastIP = r.RemoteAddr
			aTemp.OS = r.FormValue("OS")
			aTemp.Model = r.FormValue("Model")

			h = sha1.New()
			io.WriteString(h, aTemp.ID+aTemp.Name)
			hash = fmt.Sprintf("%x", h.Sum(nil))

			if _, found := a.List[hash]; found {
				aTemp = a.List[hash]
				aTemp.LastLogin = t.UTC().Format("2006-01-02 15:04:05")
				aTemp.Logins++
				a.List[hash] = aTemp
				status := "OK"
				fmt.Fprintf(w, status)
			} else {
				aTemp.FirstLogin = t.UTC().Format("2006-01-02 15:04:05")
				aTemp.Logins = 1
				aTemp.LastLogin = aTemp.FirstLogin
				a.List[hash] = aTemp
				// append new entry to file
				go a.SaveAppList()
				// return new registration Created OK
				fmt.Fprintf(w, "NEW")
			}
		}
	}
}

func postControllerIP(w http.ResponseWriter, r *http.Request) {
	loc := "(getLocalIP)"
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

/*
// CloseApp saves current log and exits the app gracefully
func CloseApp(in string, save bool) {
	if save {
		file.SaveAuxApp(auxAppList)
	}
	os.Exit(0)
}
*/
