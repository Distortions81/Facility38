//go:build js && wasm
// +build js,wasm

package main

import "syscall/js"

func generateFile(this js.Value, inputs []js.Value) interface{} {
	file := []byte("Hello, World!")
	return file
}
func sendFile() {
	c := make(chan struct{}, 0)
	println("Wasm loaded")
	js.Global().Set("generateFile", js.FuncOf(generateFile))
	<-c
}

func receiveFile(this js.Value, inputs []js.Value) interface{} {
	//file := inputs[0].Array()
	// Do something with the file, for example, store it or display it.
	return nil
}
func getFile() {
	c := make(chan struct{}, 0)
	println("Wasm loaded")
	js.Global().Set("receiveFile", js.FuncOf(receiveFile))
	<-c
}
