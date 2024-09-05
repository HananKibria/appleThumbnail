package main

import (
	"syscall/js"
	"fmt"
)

func processHeifFile(this js.Value, args []js.Value) interface{} {
	buffer := args[0]
	libheifInstance := args[1]
	callback := js.FuncOf(func(this js.Value, callbackArgs []js.Value) interface{} {
		// Handle callback arguments here if necessary
		return nil
	})

	decoder := libheifInstance.Get("HeifDecoder").New()
	image := decoder.Call("decode", buffer).Index(0)

	origWidth := image.Call("get_width").Int()
	origHeight := image.Call("get_height").Int()

	fmt.Printf("Original image size: %dx%d\n", origWidth, origHeight)
	callback.Invoke(fmt.Sprintf("Original image size: %dx%d", origWidth, origHeight))

	// Set maximum thumbnail size
	maxSize := 150
	var width, height int

	// Calculate thumbnail size while preserving aspect ratio
	if origWidth > origHeight {
		width = maxSize
		height = (origHeight * maxSize) / origWidth
	} else {
		height = maxSize
		width = (origWidth * maxSize) / origHeight
	}

	fmt.Printf("Thumbnail size: %dx%d\n", width, height)
	callback.Invoke(fmt.Sprintf("Thumbnail size: %dx%d", width, height))

	// Get the canvas and its 2D drawing context
	canvas := js.Global().Get("document").Call("getElementById", "canvas")
	if canvas.IsNull() {
		fmt.Println("Canvas element not found")
		callback.Invoke("Canvas element not found")
		return nil
	}
	ctx := canvas.Call("getContext", "2d")
	canvas.Set("width", width)
	canvas.Set("height", height)

	// Create a temporary offscreen canvas to draw the full-size image
	offscreenCanvas := js.Global().Get("document").Call("createElement", "canvas")
	offscreenCtx := offscreenCanvas.Call("getContext", "2d")
	offscreenCanvas.Set("width", origWidth)
	offscreenCanvas.Set("height", origHeight)

	// Convert the decoded image to ImageData
	imageData := offscreenCtx.Call("createImageData", origWidth, origHeight)
	image.Call("display", imageData, js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		offscreenCtx.Call("putImageData", args[0], 0, 0)

		// Now create an ImageBitmap from the ImageData
		imageBitmapPromise := js.Global().Call("createImageBitmap", args[0])
		imageBitmapPromise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			imageBitmap := args[0]

			// Draw the ImageBitmap on the offscreen canvas
			offscreenCtx.Call("drawImage", imageBitmap, 0, 0)

			// Now scale and draw the image from the offscreen canvas to the main canvas
			ctx.Call("drawImage", offscreenCanvas, 0, 0, origWidth, origHeight, 0, 0, width, height)

			// Notify that the drawing is complete
			callback.Invoke("Image processing complete")

			return nil
		}))

		return nil
	}))

	return nil
}

func main() {
	c := make(chan struct{}, 0)

	js.Global().Set("processHeifFile", js.FuncOf(processHeifFile))

	<-c
}
