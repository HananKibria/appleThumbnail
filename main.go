package main

import (
	"syscall/js"
)

func processHeifFile(this js.Value, p []js.Value) interface{} {
	buffer := p[0]          // The Uint8Array buffer containing the HEIF file data
	libheifInstance := p[1] // The libheif instance passed from JavaScript

	decoder := libheifInstance.Get("HeifDecoder").New()
	image := decoder.Call("decode", buffer).Index(0)

	width := image.Call("get_width").Int()
	height := image.Call("get_height").Int()

	canvas := js.Global().Get("document").Call("getElementById", "canvas")
	ctx := canvas.Call("getContext", "2d")
	canvas.Set("width", width)
	canvas.Set("height", height)

	imageData := ctx.Call("createImageData", width, height)

	image.Call("display", imageData, js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		ctx.Call("putImageData", args[0], 0, 0)
		return nil
	}))

	return nil
}

func main() {
	// Export the function to the global JavaScript context
	js.Global().Set("processHeifFile", js.FuncOf(processHeifFile))

	// Prevent the main Go function from exiting
	select {}
}