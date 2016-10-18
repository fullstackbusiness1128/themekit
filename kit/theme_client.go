package kit

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/themekit/theme"
)

const createThemeMaxRetries int = 3

// ThemeClient is the interactor with the shopify server. All actions are processed
// with the client.
type ThemeClient struct {
	config     Configuration
	httpClient *httpClient
	filter     eventFilter
	foreman    *foreman
}

// NewThemeClient will build a new theme client from a configuration and a theme event
// channel. The channel is used for logging all events. The configuration specifies how
// the client will behave.
func NewThemeClient(config Configuration) (ThemeClient, error) {
	httpClient, err := newHTTPClient(config)
	if err != nil {
		return ThemeClient{}, err
	}

	filter, err := newEventFilter(config.Directory, config.IgnoredFiles, config.Ignores)
	if err != nil {
		return ThemeClient{}, err
	}

	newClient := ThemeClient{
		config:     config,
		httpClient: httpClient,
		filter:     filter,
		foreman:    newForeman(newLeakyBucket(config.BucketSize, config.RefillRate, 1)),
	}

	go newClient.process()

	return newClient, nil
}

// GetConfiguration will return the clients built config. This is useful for grabbing
// things like urls and domains.
func (t ThemeClient) GetConfiguration() Configuration {
	return t.config
}

// NewFileWatcher creates a new filewatcher using the theme clients file filter
func (t ThemeClient) NewFileWatcher(notifyFile string, callback func(ThemeClient, AssetEvent, error)) error {
	return newFileWatcher(t, t.config.Directory, true, t.filter, callback)
}

// AssetList will return a slice of remote assets from the shopify servers. The
// assets are sorted and any ignored files based on your config are filtered out.
func (t ThemeClient) AssetList() ([]theme.Asset, Error) {
	resp, err := t.httpClient.AssetQuery(Retrieve, map[string]string{})
	if err != nil && err.Fatal() {
		return []theme.Asset{}, err
	}
	sort.Sort(theme.ByAsset(resp.Assets))
	return t.filter.filterAssets(ignoreCompiledAssets(resp.Assets)), err
}

// LocalAssets will return a slice of assets from the local disk. The
// assets are filtered based on your config.
func (t ThemeClient) LocalAssets() ([]theme.Asset, error) {
	dir := fmt.Sprintf("%s%s", t.config.Directory, string(filepath.Separator))
	assets, err := theme.LoadAssetsFromDirectory(dir, t.filter.matchesFilter)
	if err != nil {
		return assets, err
	}
	return assets, nil
}

// LoadAsset will load a single local asset on disk. It will return an error if there
// is a problem loading the asset.
func (t ThemeClient) LocalAsset(filename string) (theme.Asset, error) {
	return theme.LoadAsset(t.config.Directory, filename)
}

// Asset will load up a single remote asset from the remote shopify servers.
func (t ThemeClient) Asset(filename string) (theme.Asset, Error) {
	resp, err := t.httpClient.AssetQuery(Retrieve, map[string]string{"asset[key]": filename})
	if err != nil {
		return theme.Asset{}, err
	}
	return resp.Asset, nil
}

// CreateTheme will create a unpublished new theme on your shopify store and then
// return a new theme client with the configuration of the new client.
func CreateTheme(name, zipLocation string) (ThemeClient, error) {
	config, _ := NewConfiguration()
	err := config.validateNoThemeId()
	if err != nil {
		return ThemeClient{}, fmt.Errorf("Invalid options: %v", err)
	}

	client, err := NewThemeClient(config)
	if err != nil {
		return client, err
	}

	var resp *ShopifyResponse
	var respErr Error
	retries := 0
	for retries < createThemeMaxRetries {
		resp, respErr = client.httpClient.NewTheme(name, zipLocation)
		if respErr != nil {
			if respErr.Fatal() {
				return client, respErr
			}
		}

		retries++

		if resp.Successful() {
			Logf(
				"[%s]Successfully created theme '%s' with id of %s on shop %s",
				GreenText(fmt.Sprintf("%d", resp.Code)),
				BlueText(resp.Theme.Name),
				BlueText(fmt.Sprintf("%d", resp.Theme.ID)),
				YellowText(resp.Host),
			)
			break
		}

		if retries >= createThemeMaxRetries {
			return client, KitError{fmt.Errorf("Cannot create a theme. Please check log for errors.")}
		}
	}

	for !client.isDoneProcessing(resp.Theme.ID) {
		time.Sleep(250 * time.Millisecond)
	}

	client.config.ThemeID = fmt.Sprintf("%d", resp.Theme.ID)

	return client, nil
}

func (t ThemeClient) isDoneProcessing(themeID int64) bool {
	resp, err := t.httpClient.GetTheme(themeID)
	return err == nil && resp.Theme.Previewable
}

func (t ThemeClient) CreateAsset(asset theme.Asset, callback eventCallback) {
	t.UpdateAsset(asset, callback)
}

func (t ThemeClient) UpdateAsset(asset theme.Asset, callback eventCallback) {
	go func() {
		t.foreman.JobQueue <- AssetEvent{Asset: asset, Type: Update, Callback: callback}
	}()
}

func (t ThemeClient) DeleteAsset(asset theme.Asset, callback eventCallback) {
	go func() {
		t.foreman.JobQueue <- AssetEvent{Asset: asset, Type: Remove, Callback: callback}
	}()
}

func (t ThemeClient) Perform(event AssetEvent, callback eventCallback) {
	go func() {
		event.Callback = callback
		t.foreman.JobQueue <- event
	}()
}

func (t ThemeClient) process() {
	var processWaitGroup sync.WaitGroup
	for {
		job, more := <-t.foreman.WorkerQueue
		if more {
			processWaitGroup.Add(1)
			go func() {
				t.perform(job)
				processWaitGroup.Done()
			}()
		} else {
			processWaitGroup.Wait()
			return
		}
	}
}

func (t ThemeClient) perform(event AssetEvent) {
	if t.filter.matchesFilter(event.Asset.Key) {
		event.Callback(&ShopifyResponse{}, KitError{fmt.Errorf(YellowText(fmt.Sprintf("Asset %s filtered based on ignore patterns", event.Asset.Key)))})
	}
	event.Callback(t.httpClient.AssetAction(event.Type, event.Asset))
}

func ignoreCompiledAssets(assets []theme.Asset) []theme.Asset {
	newSize := 0
	results := make([]theme.Asset, len(assets))
	isCompiled := func(a theme.Asset, rest []theme.Asset) bool {
		for _, other := range rest {
			if strings.Contains(other.Key, a.Key) {
				return true
			}
		}
		return false
	}
	for index, asset := range assets {
		if !isCompiled(asset, assets[index+1:]) {
			results[newSize] = asset
			newSize++
		}
	}
	return results[:newSize]
}
