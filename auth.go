package main

import (
	"archive/zip"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

var authSite = "https://facility38.xyz:8648"

/* Contact server for version information */
func checkVersion(silent bool) bool {
	defer reportPanic("checkVersion")

	if buildTime == "Dev Build" {
		return false
	}

	// Create HTTP client with custom transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: *useLocal,
		},
	}
	client := &http.Client{Transport: transport}

	if *useLocal {
		authSite = "https://localhost:8648"
	}

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

			silenceUpdates = true
			updateVersion = newVersion

			if respParts[2] != "" {
				dlBase := strings.TrimSuffix(respParts[2], "/")
				downloadURL = ""
				switch runtime.GOOS {
				case "linux":
					downloadURL = fmt.Sprintf("%v/Facility-38-%v-linux64.zip", dlBase, newVersion)
				case "windows":
					downloadURL = fmt.Sprintf("%v/Facility-38-%v-win64.zip", dlBase, newVersion)
				case "darwin":
					//downloadURL = fmt.Sprintf("%v/Facility-38-%v-mac64.zip", dlBase, newVersion)
				default:
					//Unsupported
				}
			}

			openWindow(windows[2])
			return true
		}
	} else if respPartLen > 0 && respParts[0] == "UpToDate" {
		chat("Update server: Facility 38 is up-to-date.")
	} else {
		return false
	}

	return false
}

const downloadPathTemp = "update.tmp.exe"

func downloadBuild() bool {
	defer reportPanic("downloadBuild")

	newBuildTemp, err := os.OpenFile(downloadPathTemp, os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		chat("Unable to create update file.")
		return false
	}
	defer newBuildTemp.Close()

	/* Create HTTP client with custom transport */
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: *useLocal,
		},
	}
	client := &http.Client{Transport: transport}

	/* Attempt to fetch the URL */
	res, err := client.Get(downloadURL)
	if err != nil {
		doLog(true, "Failed to download new build from server: %v", err)
		return false
	}

	/* Read response body */
	body, err := io.ReadAll(res.Body)
	if err != nil {
		doLog(true, "Failed to read reply from server: %v", err)
		return false
	}
	res.Body.Close()

	archive, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil || archive.File == nil {
		chat("Opening update zip file failed.")
		return false
	}

	for _, file := range archive.File {
		info := file.FileInfo()
		if !info.IsDir() {
			zipFile, err := readZipFile(archive.File[0])
			if err != nil || zipFile == nil {
				chat("Reading update zip file failed.")
				return false
			}
			_, err = newBuildTemp.Write(zipFile)
			if err != nil {
				chat("Unable to write update to disk.")
				return false
			}
			newBuildTemp.Close()
			break
		}

	}

	os.Chmod(downloadPathTemp, 0760)

	gameLock.Lock()
	nukeWorld()

	doLog(true, "Relaunching.")

	pname, _ := os.Executable()
	var args []string = make([]string, 2)
	args[0] = downloadPathTemp
	args[1] = "-relaunch=" + path.Base(pname)

	process, err := os.StartProcess(downloadPathTemp, args, &os.ProcAttr{})
	if err == nil {

		// It is not clear from docs, but Realease actually detaches the process
		err = process.Release()
		if err != nil {
			fmt.Println(err.Error())
		}

	} else {
		fmt.Println(err.Error())
	}

	os.Exit(0)
	return false
}

func readZipFile(zf *zip.File) ([]byte, error) {
	defer reportPanic("readZipFile")

	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
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

	/* Create HTTP client with custom transport */
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}
	client := &http.Client{Transport: transport}

	/* Send HTTPS POST request to server */
	response, err := client.Post(authSite, "application/json", bytes.NewBuffer([]byte("CheckAuthDev:"+Secrets.P)))
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

	/* Read server response */
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	pass := string(responseBytes)

	/* Check reply */
	if pass == "GoodAuth:"+Secrets.R {
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
