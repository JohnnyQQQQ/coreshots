package cmd

import (
	"coreshots/pkg/version"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "downloads the newest version",
	Run:   update,
}

const repoURL = "https://github.com/JohnnyQQQQ/coreshots"
const stableURL = "https://raw.githubusercontent.com/JohnnyQQQQ/coreshots/master/stable.txt"

func update(cmd *cobra.Command, args []string) {
	logger := logger()
	currentBinPath, err := os.Executable()
	if err != nil {
		level.Error(logger).Log("err", err)
		return
	}
	stableVersion, err := latestVersion(stableURL)
	if err != nil {
		level.Error(logger).Log("err", err)
		return
	}

	if stableVersion == version.Current {
		logger.Log("msg", "you have already the newest version running",
			"current_version", version.Current,
			"latest_version", stableVersion)
		return
	}

	logger.Log("msg", "upgrading binary", "path", currentBinPath, "stable_version", stableVersion)

	binaryFileName := "coreshots.exe"
	shaFileName := "coreshots.exe.sha256"
	binaryURL := assetURL(stableVersion, binaryFileName)
	shaURL := assetURL(stableVersion, shaFileName)
	logger.Log("msg", "downloading binary", "binary", binaryURL, "sha256", shaURL)
	tmpDir, err := ioutil.TempDir("", "coreshot-update-*")
	if err != nil {
		level.Error(logger).Log("err", err)
		return
	}
	defer os.RemoveAll(tmpDir)
	logger.Log("msg", "created tmp dir", "path", tmpDir)
	logger.Log("msg", "donwloading file", "url", binaryURL)
	binaryFilePath := filepath.Join(tmpDir, binaryFileName)
	err = downloadFile(binaryURL, binaryFilePath)
	if err != nil {
		level.Error(logger).Log("err", err)
		return
	}
	logger.Log("msg", "donwloading file", "url", shaURL)
	shaFilePath := filepath.Join(tmpDir, shaFileName)
	err = downloadFile(shaURL, shaFilePath)
	if err != nil {
		level.Error(logger).Log("err", err)
		return
	}
	err = validateIntegrity(binaryFilePath, shaFilePath, logger)
	if err != nil {
		level.Error(logger).Log("err", err)
		return
	}
	// hack for windows, first move the old binary to the tmp dir
	// otherwise it will result in access denied on rename
	err = os.Rename(currentBinPath, filepath.Join(tmpDir, "old_version.exe"))
	if err != nil {
		level.Error(logger).Log("err", err)
		return
	}
	err = os.Rename(binaryFilePath, currentBinPath)
	if err != nil {
		level.Error(logger).Log("err", err)
		return
	}
	logger.Log("msg", "successfully updated", "new_version", stableVersion)
}

func assetURL(version, assetName string) string {
	return fmt.Sprintf("%s/releases/download/%s/%s", repoURL, version, assetName)
}

func latestVersion(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func downloadFile(fileURL, filePath string) error {
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()
	resp, err := http.Get(fileURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func validateIntegrity(binaryPath, shaPath string, logger log.Logger) error {
	f, err := os.Open(binaryPath)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	localSHA := fmt.Sprintf("%x", h.Sum(nil))
	logger.Log("msg", "calculated binary sha", "sha256", localSHA)
	b, err := os.ReadFile(shaPath)
	if err != nil {
		return err
	}
	githubSHA := strings.TrimSuffix(string(b), "\n")
	logger.Log("msg", "read sha from github", "sha256", githubSHA)
	if localSHA != githubSHA {
		return fmt.Errorf("sha is not matching got %s but expected %s", localSHA, githubSHA)
	}
	logger.Log("msg", "successfully validate the integrity")
	return nil
}
