// Self-update: download TarkovPilot.zip from the website (the same file as on
// the download page), extract TarkovPilot.exe.
// A running exe on Windows cannot be overwritten, but it CAN be renamed:
// current → TarkovPilot.exe.old, the new one in its place, launch it, exit.
// Leftover *.old files are cleaned up on the next start (CleanupOldBinary).
package updater

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// the single zip: both the download page and self-update use it
const updateZipPath = "/pilot/TarkovPilot.zip"

// CleanupOldBinary removes leftovers of the previous update (call on start)
func CleanupOldBinary() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	old := exe + ".old"
	// right after a restart the old exe may still be held by the exiting process
	for i := 0; i < 3; i++ {
		if err := os.Remove(old); err == nil || errors.Is(err, os.ErrNotExist) {
			return
		}
		time.Sleep(1 * time.Second)
	}
}

// Update downloads and installs the new version, then restarts the app.
// Returns an error if something went wrong (the current install is untouched)
func Update(host string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	// 1) download the zip into a temp file (localhost — over http, for local testing)
	scheme := "https"
	if h := strings.Split(host, ":")[0]; h == "localhost" || h == "127.0.0.1" {
		scheme = "http"
	}
	url := fmt.Sprintf("%s://%s%s", scheme, host, updateZipPath)
	httpClient := &http.Client{Timeout: 5 * time.Minute}
	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: http %d", resp.StatusCode)
	}

	tmpZip, err := os.CreateTemp("", "tarkovpilot-update-*.zip")
	if err != nil {
		return err
	}
	tmpZipPath := tmpZip.Name()
	defer os.Remove(tmpZipPath)

	if _, err := io.Copy(tmpZip, resp.Body); err != nil {
		tmpZip.Close()
		return err
	}
	tmpZip.Close()

	// 2) extract TarkovPilot.exe from the archive into a temp file next to the current exe
	//    (same drive — the rename will be atomic)
	newExePath := exe + ".new"
	if err := extractExe(tmpZipPath, newExePath); err != nil {
		return err
	}

	// 3) swap: current → .old, new → into the current's place
	oldPath := exe + ".old"
	_ = os.Remove(oldPath)
	if err := os.Rename(exe, oldPath); err != nil {
		os.Remove(newExePath)
		return err
	}
	if err := os.Rename(newExePath, exe); err != nil {
		// rollback
		_ = os.Rename(oldPath, exe)
		return err
	}

	// 4) launch the new version and exit
	cmd := exec.Command(exe, "updated")
	cmd.Dir = filepath.Dir(exe)
	if err := cmd.Start(); err != nil {
		return err
	}

	os.Exit(0)
	return nil
}

// extractExe pulls TarkovPilot.exe out of the zip
func extractExe(zipPath, dstPath string) error {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer zr.Close()

	for _, f := range zr.File {
		if strings.EqualFold(filepath.Base(f.Name), "TarkovPilot.exe") {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
			if err != nil {
				return err
			}
			defer dst.Close()

			_, err = io.Copy(dst, rc)
			return err
		}
	}
	return errors.New("TarkovPilot.exe not found in update archive")
}
