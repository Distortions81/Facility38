//go:build js && wasm
// +build js,wasm

package main

import (
	"syscall/js"
)

/* Send data to javascript */
func sendBytes(filename string, data []byte) {

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

/* Create saveGame load button, and event */
func init() {
	go func() {
		// create a file input element
		fileInput := js.Global().Get("document").Call("createElement", "input")
		fileInput.Set("type", "file")

		// add the file input element to the DOM
		js.Global().Get("document").Get("body").Call("appendChild", fileInput)

		// style the file input element to be fixed and positioned at the top-left corner of the viewport
		fileInput.Get("style").Set("position", "fixed")
		fileInput.Get("style").Set("top", "0")
		fileInput.Get("style").Set("left", "0")
		fileInput.Get("style").Set("opacity", "1")

		// attach an event listener to handle the file selection
		fileInput.Call("addEventListener", "change", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			// get the selected file
			files := fileInput.Get("files")
			if files.Length() == 0 {
				return nil
			}
			file := files.Index(0)

			// create a file reader to read the file data
			reader := js.Global().Get("FileReader").New()
			reader.Call("addEventListener", "load", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				// get the file data as a byte slice
				fileData := js.Global().Get("Uint8Array").New(args[0].Get("target").Get("result"))
				data := make([]byte, fileData.Length())
				js.CopyBytesToGo(data, fileData)

				// print the file data
				//fmt.Printf("File Data: %s\n", data)
				LoadGame(true, data)
				//Chat("File loaded.")

				return nil
			}))
			reader.Call("readAsArrayBuffer", file)

			return nil
		}))
	}()
}
