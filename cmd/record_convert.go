package cmd

import (
	"bytes"
	"coreshots/pkg/assets"
	"coreshots/pkg/compare"
	"coreshots/pkg/tools"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrg/sysfont"
	"github.com/fogleman/gg"
	"github.com/go-kit/log/level"
	"github.com/icza/mjpeg"
	"github.com/kbinani/screenshot"
	"github.com/spf13/cobra"
)

var recordConvertCmd = &cobra.Command{
	Use:   "convert [name]",
	Args:  cobra.MinimumNArgs(1),
	Short: "convert a recorded match to a video",
	Run:   recordConvert,
}

func recordConvert(cmd *cobra.Command, args []string) {
	logger := logger()
	logger.Log("msg", "converting screenshots to a video")

	savePath, err := savePath(args[0])
	if err != nil {
		level.Error(logger).Log("err", err)
		return
	}

	files, err := ioutil.ReadDir(savePath)
	if err != nil {
		level.Error(logger).Log("err", err)
		return
	}

	logger.Log("msg", "video format",
		"width", int32(screenshot.GetDisplayBounds(0).Dx()),
		"height", int32(screenshot.GetDisplayBounds(0).Dy()))

	aw, err := mjpeg.New(filepath.Join(savePath, fmt.Sprintf("%s.avi", args[0])),
		int32(screenshot.GetDisplayBounds(0).Dx()),
		int32(screenshot.GetDisplayBounds(0).Dy()),
		1)
	if err != nil {
		level.Error(logger).Log("err", err)
		return
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})
	finder := sysfont.NewFinder(nil)
	font := finder.Match("Arial")
	for _, f := range files {
		logger.Log("msg", "adding frame", "file", f.Name())
		if !strings.HasSuffix(f.Name(), ".jpg") {
			logger.Log("msg", "skipping file as it's not an image", "filename", f.Name())
			continue
		}
		data, err := ioutil.ReadFile(filepath.Join(savePath, f.Name()))
		if err != nil {
			level.Error(logger).Log("err", err)
			continue
		}
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			level.Error(logger).Log("err", err)
			continue
		}
		img, err = tools.CropImage(img, compare.WQHDOverlayMapCrop)
		if err != nil {
			level.Error(logger).Log("err", err)
			continue
		}
		dc := gg.NewContext(
			screenshot.GetDisplayBounds(0).Dx(),
			screenshot.GetDisplayBounds(0).Dy())
		dc.DrawImage(img, 0, 0)
		img, _, err = image.Decode(bytes.NewReader(assets.CoreLogo))
		if err != nil {
			level.Error(logger).Log("err", err)
			continue
		}
		dc.DrawImage(img, 0, 0)
		dc.SetColor(color.White)
		if err := dc.LoadFontFace(font.Filename, 100); err != nil {
			level.Error(logger).Log("err", err)
			continue
		}
		dc.DrawStringWrapped(
			strings.ReplaceAll(f.Name()[:len(f.Name())-4], "_", ":"),
			float64(screenshot.GetDisplayBounds(0).Dx()/2),
			50,
			0.5,
			0.5,
			10000,
			1.5,
			gg.AlignCenter)

		buf := new(bytes.Buffer)
		if err := jpeg.Encode(buf, dc.Image(), nil); err != nil {
			level.Error(logger).Log("err", err)
			continue
		}
		err = aw.AddFrame(buf.Bytes())
		if err != nil {
			level.Error(logger).Log("err", err)
			continue
		}
	}
	err = aw.Close()
	if err != nil {
		level.Error(logger).Log("err", err)
	}
}
