package kit

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	debounceTimeout = 1100 * time.Millisecond
)

// FileEventCallback is the callback that is called when there is an event from
// a file watcher.
type FileEventCallback func(ThemeClient, Asset, EventType)

// FileWatcher is the object used to watch files for change and notify on any events,
// these events can then be passed along to kit to be sent to shopify.
type FileWatcher struct {
	done          chan bool
	client        ThemeClient
	mainWatcher   *fsnotify.Watcher
	reloadSignal  chan bool
	configWatcher *fsnotify.Watcher
	filter        fileFilter
	callback      FileEventCallback
	notify        string
}

func newFileWatcher(client ThemeClient, dir, notifyFile string, recur bool, filter fileFilter, callback FileEventCallback) (*FileWatcher, error) {
	mainWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	configWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	newWatcher := &FileWatcher{
		done:          make(chan bool),
		client:        client,
		mainWatcher:   mainWatcher,
		configWatcher: configWatcher,
		callback:      callback,
		filter:        filter,
	}

	go newWatcher.watchFsEvents(notifyFile)

	return newWatcher, newWatcher.watchDirectory(dir)
}

func (watcher *FileWatcher) watchDirectory(root string) error {
	root = filepath.Clean(root)
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && !watcher.filter.matchesFilter(path) && path != root {
			if err := watcher.mainWatcher.Add(path); err != nil {
				return fmt.Errorf("Could not watch directory %s: %s", path, err)
			}
		}
		return nil
	})
}

func (watcher *FileWatcher) watchFsEvents(notifyFile string) {
	var eventLock sync.Mutex
	notifyProcessed := false
	recordedEvents := map[string]chan fsnotify.Event{}

	for {
		select {
		case configEvent := <-watcher.configWatcher.Events:
			if configEvent.Op != fsnotify.Chmod {
				close(watcher.done)
				if watcher.reloadSignal != nil {
					watcher.reloadSignal <- true
				}
				return
			}
		case currentEvent, more := <-watcher.mainWatcher.Events:
			if !more {
				close(watcher.done)
				return
			}

			if currentEvent.Op == fsnotify.Chmod || watcher.filter.matchesFilter(currentEvent.Name) {
				break
			}

			eventLock.Lock()
			if _, ok := recordedEvents[currentEvent.Name]; !ok {
				recordedEvents[currentEvent.Name] = make(chan fsnotify.Event)

				go func(eventChan chan fsnotify.Event, eventName string) {
					var event fsnotify.Event

					for {
						select {
						case event = <-eventChan:
						case <-time.After(debounceTimeout):
							go handleEvent(watcher, event)
							eventLock.Lock()
							delete(recordedEvents, eventName)
							eventLock.Unlock()
							return
						}
					}
				}(recordedEvents[currentEvent.Name], currentEvent.Name)
			}
			recordedEvents[currentEvent.Name] <- currentEvent
			notifyProcessed = true
			eventLock.Unlock()
		case <-time.Tick(1 * time.Second):
			eventLock.Lock()
			if notifyProcessed && len(recordedEvents) == 0 && notifyFile != "" {
				os.Create(notifyFile)
				os.Chtimes(notifyFile, time.Now(), time.Now())
				notifyProcessed = false
			}
			eventLock.Unlock()
		}
	}
}

// WatchConfig adds a priority watcher for the config file. A true will be sent down
// the channel to notify you about a config file change. This is useful to keep
// track of version control changes
func (watcher *FileWatcher) WatchConfig(configFile string, reloadSignal chan bool) error {
	if err := watcher.configWatcher.Add(configFile); err != nil {
		return err
	}
	watcher.reloadSignal = reloadSignal
	return nil
}

// IsWatching will return true if the watcher is currently watching for file changes.
// it will return false if it has been stopped
func (watcher *FileWatcher) IsWatching() bool {
	select {
	case _, ok := <-watcher.done:
		return ok
	default:
		return true
	}
}

// StopWatching will stop the Filewatcher from watching it's directories and clean
// up any go routines doing work.
func (watcher *FileWatcher) StopWatching() {
	watcher.mainWatcher.Close()
}

func handleEvent(watcher *FileWatcher, event fsnotify.Event) {
	if !watcher.IsWatching() {
		return
	}
	eventType := Update
	if event.Op&fsnotify.Remove == fsnotify.Remove {
		eventType = Remove
	}
	asset, _ := loadAsset(filepath.Dir(event.Name), filepath.Base(event.Name))
	watcher.callback(watcher.client, asset, eventType)
}
