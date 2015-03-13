package commands

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/csaunders/phoenix"
	"os"
	"path/filepath"
)

func DownloadCommand(args map[string]interface{}) chan bool {
	return toClientAndFilesAsync(args, Download)
}

func Download(client phoenix.ThemeClient, filenames []string) (done chan bool) {
	done = make(chan bool)
	eventLog := make(chan phoenix.ThemeEvent)

	if len(filenames) <= 0 {
		assets, errs := client.AssetList()
		go drainErrors(errs)
		go downloadAllFiles(assets, done, eventLog)
	} else {
		go downloadFiles(client.Asset, filenames, done, eventLog)
	}

	return done
}

func downloadAllFiles(assets chan phoenix.Asset, done chan bool, eventLog chan phoenix.ThemeEvent) {
	for {
		asset, more := <-assets
		if more {
			writeToDisk(asset, eventLog)
		} else {
			done <- true
			return
		}
	}
}

func downloadFiles(retrievalFunction phoenix.AssetRetrieval, filenames []string, done chan bool, eventLog chan phoenix.ThemeEvent) {
	for _, filename := range filenames {
		if asset, err := retrievalFunction(filename); err != nil {
			phoenix.NotifyError(err)
		} else {
			writeToDisk(asset, eventLog)
		}
	}
	done <- true
	return
}

func writeToDisk(asset phoenix.Asset, eventLog chan phoenix.ThemeEvent) {
	dir, err := os.Getwd()
	if err != nil {
		phoenix.NotifyError(err)
		return
	}

	perms, err := os.Stat(dir)
	if err != nil {
		phoenix.NotifyError(err)
		return
	}

	filename := fmt.Sprintf("%s/%s", dir, asset.Key)
	err = os.MkdirAll(filepath.Dir(filename), perms.Mode())
	if err != nil {
		phoenix.NotifyError(err)
		return
	}

	file, err := os.Create(filename)
	if err != nil {
		phoenix.NotifyError(err)
		return
	}
	defer file.Sync()
	defer file.Close()

	var data []byte
	switch {
	case len(asset.Value) > 0:
		data = []byte(asset.Value)
	case len(asset.Attachment) > 0:
		data, err = base64.StdEncoding.DecodeString(asset.Attachment)
		if err != nil {
			phoenix.NotifyError(errors.New(fmt.Sprintf("Could not decode %s. error: %s", asset.Key, err)))
			return
		}
	}

	if len(data) > 0 {
		_, err = file.Write(data)
	}

	if err != nil {
		phoenix.NotifyError(err)
	} else {
		event := basicEvent{
			Title:     "FS Event",
			EventType: "Write",
			Target:    filename,
			etype:     "fsevent",
			Formatter: func(b basicEvent) string {
				return fmt.Sprintf("Successfully wrote %s to disk", b.Target)
			},
		}
		logEvent(event, eventLog)
	}
}
