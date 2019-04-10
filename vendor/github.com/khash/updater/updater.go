package updater

import (
	"encoding/json"
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

// Updater is responsible for checking for updates and updating the running executable
type Updater struct {
	CurrentVersion string

	currentVersion *version.Version
	options        *Options
}

// NewUpdater returns a new updater instance
func NewUpdater(currentVersion string, options *Options) (*Updater, error) {
	if options.RemoteURL == "" {
		panic("no RemoteURL")
	}
	if !strings.HasSuffix(options.RemoteURL, "/") {
		options.RemoteURL = options.RemoteURL + "/"
	}
	if options.VersionSpecsFilename == "" {
		options.VersionSpecsFilename = "versions.json"
	}
	if options.Channel == "" {
		options.Channel = "dev"
	}
	if options.BinPattern == "" {
		options.BinPattern = "{{OS}}_{{ARCH}}_{{VERSION}}"
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

// Run runs the updater
func (u *Updater) Run() error {
	remoteVersion, err := u.getRemoteVersion()
	if err != nil {
		return err
	}

	if !u.options.Silent {
		fmt.Printf("Local Version %v - Remote Version: %v\n", u.currentVersion, remoteVersion)
	}
	if u.currentVersion.LessThan(remoteVersion) {
		err = u.downloadAndReplace(remoteVersion)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *Updater) downloadAndReplace(remoteVersion *version.Version) error {
	// fetch the new file
	bodyResp, err := http.Get(generateURL(u.options.BinURL(), remoteVersion.String()))
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

	return nil
}

func (u *Updater) getRemoteVersion() (*version.Version, error) {
	response, err := http.Get(u.options.VersionSpecsURL())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// get the version descriptor
	var versions VersionSpecs
	err = json.Unmarshal(b, &versions)
	if err != nil {
		return nil, err
	}

	v, err := versions.GetVersionByChannel(u.options.Channel)
	if err != nil {
		return nil, err
	}

	remoteVersion, err := version.NewVersion(v.Version)
	if err != nil {
		return nil, fmt.Errorf("remote version is '%s'. %s", string(b), err)
	}

	return remoteVersion, nil
}

func generateURL(path string, version string) string {
	path = strings.Replace(path, "{{OS}}", runtime.GOOS, -1)
	path = strings.Replace(path, "{{ARCH}}", runtime.GOARCH, -1)
	path = strings.Replace(path, "{{VERSION}}", version, -1)

	return path
}
