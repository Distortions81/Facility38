//go:build !js && !wasm
// +build !js,!wasm

package main

/*
 * This file is a stub for non-wasm platforms.
 * Without it, some functions would have no
 * reference and the compile would fail
 */

func sendBytes(name string, bytes []byte) {

}
