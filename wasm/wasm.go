//go:build js && wasm
// +build js,wasm

package wasm

import "syscall/js"

func SendBytes(string name, bytes []byte) {
	js.Global().Call("sendBytes", bytes)
}

//export sendBytesToJS
func sendBytesToJS(ptr unsafe.Pointer, len int) {
	bytes := make([]byte, len)
	copy(bytes, (*[1 << 30]byte)(ptr)[:len:len])
	SendBytes(bytes)
}
