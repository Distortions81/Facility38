//go:build js && wasm
// +build js,wasm

package wasm

import "syscall/js"

func SendBytes(filename string, data []byte) {

	// convert the Go byte slice to a JavaScript Uint8Array object
	jsData := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(jsData, data)

	// create a new JavaScript Blob object from the Uint8Array and filename
	jsBlob := js.Global().Get("Blob").New([]interface{}{jsData}, map[string]interface{}{
		"type": "application/octet-stream",
	})

	// create a new URL object from the Blob
	jsUrl := js.Global().Get("URL").Call("createObjectURL", jsBlob)

	// create a new download link
	jsDownloadLink := js.Global().Get("document").Call("createElement", "a")
	jsDownloadLink.Set("download", filename)
	jsDownloadLink.Set("href", jsUrl)
	js.Global().Get("document").Get("body").Call("appendChild", jsDownloadLink)

	// simulate a click on the download link to initiate the download
	jsDownloadLink.Call("click")

	// remove the download link from the DOM
	js.Global().Get("document").Get("body").Call("removeChild", jsDownloadLink)
}
