package main

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
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/adrg/sysfont"
	"github.com/fogleman/gg"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/icza/mjpeg"
	"github.com/kbinani/screenshot"
)

const screenshotInterval = time.Second * 10

func main() {
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	logger.Log("msg", "starting to capture screenshots")

	if len(os.Args) == 1 || os.Args[1] == "" {
		level.Error(logger).Log("err", "no name provided, usage ./coreshoot NAME_OF_THE_RECORDING")
		return
	}

	userDir, err := os.UserHomeDir()
	if err != nil {
		level.Error(logger).Log("err", err)
		return
	}

	savingDir := filepath.Join(userDir, "coreshoot", os.Args[1])
	if err := os.MkdirAll(savingDir, os.ModePerm); err != nil {
		level.Error(logger).Log("err", err)
		return
	}
	// setup proper cleanup after the program does exit
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		createVideo(logger, savingDir)
		os.Exit(0)
	}()
	// take a screenshot every n seconds
	for {
		bounds := screenshot.GetDisplayBounds(0)

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
		file, _ := os.Create(filepath.Join(savingDir, fileName))
		jpeg.Encode(file, img, nil)
		_ = file.Close()
		logger.Log("filename", fileName, "score", score, "bounds", bounds)

		time.Sleep(screenshotInterval)
	}
}

func createVideo(logger log.Logger, path string) {
	logger.Log("msg", "creating video")
	files, err := ioutil.ReadDir(path)
	if err != nil {
		level.Error(logger).Log("err", err)
		return
	}
	logger.Log("msg", "starting video", "width", int32(screenshot.GetDisplayBounds(0).Dx()), "height", int32(screenshot.GetDisplayBounds(0).Dy()))
	aw, err := mjpeg.New(filepath.Join(path, fmt.Sprintf("%s.avi", os.Args[1])),
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
		data, err := ioutil.ReadFile(filepath.Join(path, f.Name()))
		if err != nil {
			level.Error(logger).Log("err", err)
			continue
		}
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			level.Error(logger).Log("err", err)
			continue
		}
		img, err = tools.CropImage(img, compare.WQHDMapCrop)
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
