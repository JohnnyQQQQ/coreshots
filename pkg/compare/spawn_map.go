package compare

import (
	_ "embed"
	"image"
)

var WQHDSpawnMapSniffCrop = image.Rect(1000, 1440-300, 1600, 1440)
var WQHDSpawnMapCrop = image.Rect(813, 201, 1747, 1139)

//go:embed samples/spawn_map_1.jpg
var seedBytesSpawnMap []byte
