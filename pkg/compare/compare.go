package compare

import (
	"bytes"
	"coreshoots/pkg/tools"
	_ "embed"
	"image"
	"log"

	"github.com/vitali-fedulov/images4"
)

const threshold = 500000

type MapType string

const (
	OverlayMap MapType = "overlay"
	SpawnMap   MapType = "spawn"
)

var seedIconSpawnMap images4.IconT
var seedIconOverlayMap images4.IconT

func init() {

	img, _, err := image.Decode(bytes.NewReader(seedBytesSpawnMap))
	if err != nil {
		log.Fatal(err)
	}
	croppedImage, err := tools.CropImage(img, WQHDSpawnMapSniffCrop)
	if err != nil {
		log.Fatal(err)
	}
	seedIconSpawnMap = images4.Icon(croppedImage)

	img, _, err = image.Decode(bytes.NewReader(seedBytesOverlayMap))
	if err != nil {
		log.Fatal(err)
	}
	croppedImage, err = tools.CropImage(img, WQHDOverlayMapSniffCrop)
	if err != nil {
		log.Fatal(err)
	}
	seedIconOverlayMap = images4.Icon(croppedImage)
}

func IsValidImage(img image.Image, mapType MapType) (bool, int, error) {
	croppedImage, err := tools.CropImage(img, getCropRect(mapType))
	if err != nil {
		return false, 0, err
	}
	icon := images4.Icon(croppedImage)
	var m1, m2, m3 = 0.0, 0.0, 0.0
	switch mapType {
	case OverlayMap:
		m1, m2, m3 = images4.EucMetric(seedIconOverlayMap, icon)
	case SpawnMap:
		m1, m2, m3 = images4.EucMetric(seedIconSpawnMap, icon)
	}
	score := int(m1 + m2 + m3)
	return score <= threshold, score, nil
}

func getCropRect(mod MapType) image.Rectangle {
	switch mod {
	case OverlayMap:
		return WQHDOverlayMapSniffCrop
	case SpawnMap:
		return WQHDSpawnMapSniffCrop
	}
	return WQHDOverlayMapSniffCrop
}
