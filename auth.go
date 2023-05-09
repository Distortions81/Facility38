package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"
)

const authSite = "https://facility38.xyz:8648"

/* Contact server for version information */
func checkVersion(silent bool) bool {
	defer reportPanic("checkVersion")

	if buildTime == "Dev Build" {
		return false
	}

	// Create HTTP client with custom transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}
	client := &http.Client{Transport: transport}

	postString := fmt.Sprintf("CheckUpdateDev:v%03v-%v\n", version, buildTime)
	// Send HTTPS POST request to server
	response, err := client.Post(authSite, "application/json", bytes.NewBuffer([]byte(postString)))
	if err != nil {
		txt := "Unable to connect to update server."
		chat(txt)
		statusText = txt
		return false
	}
	defer response.Body.Close()

	// Read server response
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	/* Parse reply */
	resp := string(responseBytes)
	respParts := strings.Split(resp, "\n")
	respPartLen := len(respParts)

	var newVersion string
	//var dlURL string

	if respPartLen > 2 {
		if respParts[0] == "Update" {
			newVersion = respParts[1]
			//dlURL = respParts[2]

			if wasmMode {
				go chatDetailed("The game is out of date.\nYou may need to refresh your browser.", ColorOrange, 60*time.Second)
				return true
			}

			buf := fmt.Sprintf("New version available: %v", newVersion)
			if respParts[2] != "" {
				dlBase := strings.TrimSuffix(respParts[2], "/")
				downloadURL := ""
				switch runtime.GOOS {
				case "linux":
					downloadURL = fmt.Sprintf("%v/Facility38-%v-linux64.zip", dlBase, newVersion)
				case "windows":
					downloadURL = fmt.Sprintf("%v/Facility38-%v-win64.zip", dlBase, newVersion)
				case "darwin":
					//downloadURL = fmt.Sprintf("%v/Facility38-%v-mac64.zip", dlBase, newVersion)
				default:
					//Unsupported
				}

				/* TODO Open dialog box with prompt to auto-update and progress bar */
				if downloadURL != "" {
					//downloadBuild(downloadURL)
				} else {
					chat(downloadURL)
				}

			}
			silenceUpdates = true
			chatDetailed(buf, color.White, 60*time.Second)
			return true
		}
	} else if respPartLen > 0 && respParts[0] == "UpToDate" {
		chat("Update server: Facility 38 is up-to-date.")
	} else {
		return false
	}

	return false
}

func downloadBuild(downloadURL string) bool {
	/* Attempt to fetch the URL */
	res, err := http.Get(downloadURL)
	if err != nil {
		doLog(true, "Failed to download new build from server: %v", err)
		return false
	}

	/* Read response body */
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		doLog(true, "Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		doLog(true, "Failed to read reply from server: %v", err)
	}

	return false
}

/* Check server for authorization information */
func checkAuth() bool {
	defer reportPanic("checkAuth")

	if buildTime == "Dev Build" {
		authorized.Store(true)
		return true
	}

	good := loadSecrets()
	if !good {
		chat("Key load failed.")
		return false
	}

	// Create HTTP client with custom transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}
	client := &http.Client{Transport: transport}

	// Send HTTPS POST request to server
	response, err := client.Post(authSite, "application/json", bytes.NewBuffer([]byte("CheckAuthDev:"+Secrets[0].P)))
	if err != nil {
		txt := "Unable to connect to auth server."
		chat(txt)
		authorized.Store(false)
		statusText = txt

		/* Sleep for a bit, and try again */
		time.Sleep(time.Second * 10)
		go checkAuth()

		return false
	}
	defer response.Body.Close()

	// Read server response
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	pass := string(responseBytes)

	/* Check reply */
	if pass == "GoodAuth:"+Secrets[0].R {
		//Chat("Auth server approved! Have fun!")
		authorized.Store(true)
		return true
	}

	/* Server said we are no-go */
	txt := "Auth server did not approve."
	chat(txt)
	authorized.Store(false)
	statusText = txt
	return false
}
