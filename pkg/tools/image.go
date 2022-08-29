package tools

import (
	"fmt"
	"image"
)

type subImager interface {
	SubImage(r image.Rectangle) image.Image
}

func CropImage(img image.Image, cropRect image.Rectangle) (image.Image, error) {
	simg, ok := img.(subImager)
	if !ok {
		return nil, fmt.Errorf("image does not support cropping")
	}
	return simg.SubImage(cropRect), nil
}
