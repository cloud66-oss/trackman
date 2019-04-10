package updater

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/mitchellh/ioprogress"
)

// Options defines options used in an Updater instance
type Options struct {
	VersionURL string
	BinURL     string
	Silent bool
}

// Updater is responsible for checking for updates and updating the running executable
type Updater struct {
	CurrentVersion string

	currentVersion *version.Version
	options        *Options
}

func NewUpdater(currentVersion string, options *Options) (*Updater, error) {
	if options.VersionURL == "" {
		panic("no VersionURL")
	}
	if options.BinURL == "" {
		panic("no BindURL")
	}

	v, err := version.NewVersion(currentVersion)
	if err != nil {
		return nil, err
	}

	return &Updater{
		currentVersion: v,
		options:        options,
	}, nil
}

func generateURL(path string, version string) string {
	path = strings.Replace(path, "{{OS}}", runtime.GOOS, -1)
	path = strings.Replace(path, "{{ARCH}}", runtime.GOARCH, -1)
	path = strings.Replace(path, "{{VERSION}}", version, -1)

	return path
}

func (u *Updater) RunAutoUpdater() error {
	response, err := http.Get(u.options.VersionURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	remoteVersion, err := version.NewVersion(string(b))
	if err != nil {
		return fmt.Errorf("remote version is '%s'. %s", string(b), err)
	}

	if !u.options.Silent {
		fmt.Printf("Local Version %v - Remote Version: %v\n", u.currentVersion, remoteVersion)
	}
	if u.currentVersion.LessThan(remoteVersion) {
		// fetch the new file
		bodyResp, err := http.Get(generateURL(u.options.BinURL, remoteVersion.String()))
		if err != nil {
			return err
		}
		defer bodyResp.Body.Close()

		progressR := &ioprogress.Reader{
			Reader:       bodyResp.Body,
			Size:         bodyResp.ContentLength,
			DrawInterval: 500 * time.Millisecond,
			DrawFunc: ioprogress.DrawTerminalf(os.Stdout, func(progress, total int64) string {
				bar := ioprogress.DrawTextFormatBar(40)
				return fmt.Sprintf("%s %20s", bar(progress, total), ioprogress.DrawTextFormatBytes(progress, total))
			}),
		}

		var data []byte
		if !u.options.Silent {
			data, err = ioutil.ReadAll(progressR)
			if err != nil {
				return err
			}
		} else {
			data, err = ioutil.ReadAll(bodyResp.Body)
			if err != nil {
				return err
			}
		}

		dest, err := os.Executable()
		if err != nil {
			return err
		}

		// Move the old version to a backup path that we can recover from
		// in case the upgrade fails
		destBackup := dest + ".bak"
		if _, err := os.Stat(dest); err == nil {
			rErr := os.Rename(dest, destBackup)
			if rErr != nil {
				fmt.Println(rErr)
			}
		}

		if !u.options.Silent {
			fmt.Printf("Downloading the new version to %s\n", dest)
		}

		if err := ioutil.WriteFile(dest, data, 0755); err != nil {
			rErr := os.Rename(destBackup, dest)
			if rErr != nil {
				fmt.Println(rErr)
			}

			return err
		}

		// Removing backup
		rErr := os.Remove(destBackup)
		if rErr != nil {
			fmt.Println(rErr)
		}

		os.Exit(0)
	}

	return nil
}

