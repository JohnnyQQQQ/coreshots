package cmd

import (
	"bytes"
	"coreshoots/pkg/assets"
	"coreshoots/pkg/compare"
	"coreshoots/pkg/tools"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adrg/sysfont"
	"github.com/fogleman/gg"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/icza/mjpeg"
	"github.com/kbinani/screenshot"
	"github.com/spf13/cobra"
)

const screenshotInterval = time.Second * 10

func init() {
	recordCmd.AddCommand(recordStartCmd, recordConvertCmd)
}

var recordCmd = &cobra.Command{
	Use: "record",
	Run: func(cmd *cobra.Command, args []string) {},
}

var recordStartCmd = &cobra.Command{
	Use:   "start [name]",
	Args:  cobra.MinimumNArgs(1),
	Short: "starts a new recording of a match",

	Run: recordMatch,
}
var recordConvertCmd = &cobra.Command{
	Use:   "convert [name]",
	Args:  cobra.MinimumNArgs(1),
	Short: "convert a recorded match to a video",
	Run:   convertRecording,
}

func recordMatch(cmd *cobra.Command, args []string) {
	logger := logger()
	logger.Log("msg", "starting to capture screenshots")

	savePath, err := savePath(args[0])
	if err != nil {
		level.Error(logger).Log("err", err)
		return
	}
	if err := os.MkdirAll(savePath, os.ModePerm); err != nil {
		level.Error(logger).Log("err", err)
		return
	}
	logger.Log("save_path", savePath)

	bounds := screenshot.GetDisplayBounds(0)
	if bounds.Dx() != 2560 && bounds.Dy() != 1440 {
		level.Error(logger).Log("err", fmt.Errorf("screen is not WQHD"), "width",
			bounds.Dx(), "height", bounds.Dy())
		return
	}

	// take a screenshot every n seconds
	for {

		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			panic(err)
		}

		isValid, score, err := compare.IsValidImage(img, compare.OverlayMap)
		if !isValid {
			if err != nil {
				level.Error(logger).Log("err", err)
			}
			continue
		}

		fileName := fmt.Sprintf("%s.jpg", time.Now().Format("15_04_05"))
		file, _ := os.Create(filepath.Join(savePath, fileName))
		jpeg.Encode(file, img, nil)
		_ = file.Close()
		logger.Log("filename", fileName, "score", score, "bounds", bounds)

		time.Sleep(screenshotInterval)
	}
}

func convertRecording(cmd *cobra.Command, args []string) {
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
	logger.Log("msg", "video format", "width", int32(screenshot.GetDisplayBounds(0).Dx()), "height", int32(screenshot.GetDisplayBounds(0).Dy()))
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

func logger() log.Logger {
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	return log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
}

func savePath(name string) (string, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userDir, "coreshots", name), nil
}
