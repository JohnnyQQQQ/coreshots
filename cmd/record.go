package cmd

import (
	"coreshots/pkg/compare"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-kit/log"
	"github.com/spf13/cobra"
)

const screenshotInterval = time.Second * 10

var mod string

func init() {
	recordCmd.PersistentFlags().StringVarP(&mod, "mod", "m", string(compare.SpawnMap),
		fmt.Sprintf("either \"%s\" or \"%s\"", compare.SpawnMap, compare.OverlayMap))
	recordCmd.AddCommand(recordStartCmd, recordConvertCmd)
}

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Start a new recording or convert an existing one to a video",
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
