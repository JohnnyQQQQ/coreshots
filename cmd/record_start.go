package cmd

import (
	"coreshots/pkg/compare"
	"fmt"
	"image/jpeg"
	"os"
	"path/filepath"
	"time"

	"github.com/go-kit/log/level"
	"github.com/kbinani/screenshot"
	"github.com/spf13/cobra"
)

var recordStartCmd = &cobra.Command{
	Use:   "start [name]",
	Args:  cobra.MinimumNArgs(1),
	Short: "starts a new recording of a match",
	Run:   recordStart,
}

func recordStart(cmd *cobra.Command, args []string) {
	logger := logger()
	logger.Log("msg", "starting to capture screenshots")

	if mod != string(compare.OverlayMap) && mod != string(compare.SpawnMap) {
		level.Error(logger).Log("err", fmt.Errorf("mod flag has an invalid value, only %s and %s are accepted",
			string(compare.SpawnMap), string(compare.OverlayMap)), "value", mod)
		return
	}

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

	valid, nonValid := 0, 0

	// take a screenshot every n seconds
	for {

		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			panic(err)
		}

		isValid, score, err := compare.IsValidImage(img, compare.MapType(mod))
		if !isValid {
			if err != nil {
				level.Error(logger).Log("err", err)
			}
			nonValid++
			logger.Log("msg", "screenshot was not valid", "non_valid_count", nonValid, "valid_count", valid)
			continue
		}

		fileName := fmt.Sprintf("%s.jpg", time.Now().Format("15_04_05"))
		file, _ := os.Create(filepath.Join(savePath, fileName))
		jpeg.Encode(file, img, nil)
		_ = file.Close()
		valid++
		logger.Log("msg", "screenshot was valid", "non_valid_count", nonValid, "valid_count", valid, "filename", fileName, "score", score, "bounds", bounds)
		time.Sleep(screenshotInterval)
	}
}
