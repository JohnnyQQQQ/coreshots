package compare

import (
	_ "embed"
	"image"
)

var WQHDOverlayMapSniffCrop = image.Rect(1650, 1130, 1842, 1268)
var WQHDOverlayMapCrop = image.Rect(695, 97, 1865, 1293)

//go:embed samples/overlay_map_1.jpg
var seedBytesOverlayMap []byte
