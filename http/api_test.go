package api

import (
	"fmt"
	"strings"
	"testing"
)

func TestGetServerIP(t *testing.T) {
	const testFunc = "getServerIP..."
	fmt.Printf("Testing: %s\n", testFunc)

	var c configT
	const local = "127.0.0.1"

	tt := []struct {
		name  string
		local bool
		tlsOK bool
	}{
		{name: "run local, no tls loaded", local: true, tlsOK: false},
		{name: "run local, with tls loaded", local: true, tlsOK: true},
		{name: "run WAN, no tls loaded", local: false, tlsOK: false},
		{name: "run WAN, with tls loaded", local: false, tlsOK: true},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			port := "80"
			if tc.tlsOK {
				port = "443"
			}
			if tc.local {
				port = "8080"
			}
			c.Local = tc.local
			c.getServerIP(tc.tlsOK)
			ip := strings.Split(c.serverIP, ":")
			if ip[1] != port {
				t.Fatalf("Expected port %s, got %s\n", port, ip[1])
			}
			if tc.local && ip[0] != local {
				t.Fatalf("Expected loopback IP, got %s\n", ip[0])
			}
			if !tc.local && ip[0] == local {
				t.Fatalf("Expected non loopback IP, got %s\n", ip[0])
			}
		})
	}
}

func TestGetMode(t *testing.T) {
	const testFunc = "getMode..."
	fmt.Printf("Testing: %s\n", testFunc)

	tt := []struct {
		name     string
		lastMode string
		lastEnum modeT
	}{
		{name: "lastMode: clientSP", lastMode: "clientSP", lastEnum: clientSP},
		{name: "lastMode: clientComp", lastMode: "clientComp", lastEnum: clientComp},
		{name: "lastMode: clientIOT", lastMode: "clientIOT", lastEnum: clientIOT},
		{name: "lastMode: ctrlLite", lastMode: "ctrlLite", lastEnum: ctrlLite},
		{name: "lastMode: ctrlPro", lastMode: "ctrlPro", lastEnum: ctrlPro},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			mode := getMode(tc.lastMode)
			if mode != tc.lastEnum {
				t.Fatalf("Expected %v, got %v\n", tc.lastEnum, mode)
			}
		})
	}
}

func TestCompareFirstThreeIPDigits(t *testing.T) {
	const testFunc = "compareFirstThreeIPDigits..."
	fmt.Printf("Testing: %s\n", testFunc)

	var appList appListT
	var ai appInfoT
	var c ipT

	ai.ID = "testID"
	ai.LastPrivateIP = "123.123.123.123"

	appList = make(map[string]appInfoT)
	appList[ai.ID] = ai

	tt := []struct {
		name    string
		index   string
		private string
		foundIP bool
		out     string
	}{
		{
			name:    "matching private IP",
			index:   "testID",
			private: "123.123.123.123",
			foundIP: true,
			out:     "OK:" + appList["testID"].LastPrivateIP,
		},
		{
			name:    "non-matching private IP",
			index:   "testID",
			private: "124.124.124.124",
			foundIP: false,
			out:     "ERROR:Not Found",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			c.private = tc.private
			msg, found := appList.compareFirstThreeIPDigits(tc.index, c)
			if found != tc.foundIP {
				t.Fatalf("Expected %v, got %v\nMessage: %s, got %s\n", tc.foundIP, found, tc.out, msg)
			}
		})
	}
}
