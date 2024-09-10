package main

import (
	"syscall/js"
	"fmt"
)

func processHeifFile(this js.Value, args []js.Value) interface{} {
	fmt.Println("processHeifFile called")
	buffer := args[0]
	libheifInstance := args[1]
	callback := args[2]

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

	// Create a temporary offscreen canvas to draw the thumbnail
	offscreenCanvas := js.Global().Get("document").Call("createElement", "canvas")
	offscreenCtx := offscreenCanvas.Call("getContext", "2d")
	offscreenCanvas.Set("width", width)
	offscreenCanvas.Set("height", height)

	// Convert the decoded image to ImageData
	imageData := offscreenCtx.Call("createImageData", origWidth, origHeight)
	image.Call("display", imageData, js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		offscreenCtx.Call("putImageData", args[0], 0, 0)

		// Now create an ImageBitmap from the ImageData
		imageBitmapPromise := js.Global().Call("createImageBitmap", args[0])
		imageBitmapPromise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			imageBitmap := args[0]

			// Draw the ImageBitmap on the offscreen canvas
			offscreenCtx.Call("drawImage", imageBitmap, 0, 0, origWidth, origHeight, 0, 0, width, height)

			// Create a Blob from the offscreen canvas
			offscreenCanvas.Call("toBlob", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				blob := args[0]

				// Generate a URL for the Blob
				url := js.Global().Get("URL").Call("createObjectURL", blob)

				// Return the URL back to the JavaScript code
				fmt.Println("Returning URL:", url.String())
				callback.Invoke(url.String())

				return nil
			}), "image/png")

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
