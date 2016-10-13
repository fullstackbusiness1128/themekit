package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/Shopify/themekit/kit"
)

const (
	banner                 string = "----------------------------------------"
	updateAvailableMessage string = `
| An update for Theme Kit is available |
|                                      |
| To apply the update simply type      |
| the following command:               |
|                                      |
| theme update                         |`
)

var (
	environments kit.Environments
	themeClients []kit.ThemeClient
	directory    string
	configPath   string
	environment  string
	allenvs      bool
	notifyFile   string
	password     string
	themeid      string
	domain       string
	bucketsize   int
	refillrate   int
	concurrency  int
	proxy        string
	timeout      time.Duration

	bootstrapVersion string
	bootstrapPrefix  string
	setThemeID       bool
)

var ThemeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Theme Kit is a tool kit for manipulating shopify themes",
	Long: `Theme Kit is a tool kit for manipulating shopify themes

Theme Kit is a Fast and cross platform tool that enables you
to build shopify themes with ease.

Complete documentation is available at http://themekit.cat`,
}

func init() {
	pwd, _ := os.Getwd()
	configPath = filepath.Join(pwd, "config.yml")

	ThemeCmd.PersistentFlags().StringVarP(&configPath, "config", "c", configPath, "path to config.yml")
	ThemeCmd.PersistentFlags().StringVarP(&environment, "env", "e", kit.DefaultEnvironment, "envionment to run the command")

	ThemeCmd.PersistentFlags().StringVarP(&directory, "dir", "d", "", "directory that command will take effect. (default current directory)")
	ThemeCmd.PersistentFlags().StringVar(&password, "password", "", "theme password. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().StringVar(&themeid, "themeid", "", "theme id. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().StringVar(&domain, "domain", "", "your shopify domain. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().IntVar(&bucketsize, "bucket", 0, "the bucket size for throttling. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().IntVar(&refillrate, "refill", 0, "the refill rate for throttling. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().IntVar(&concurrency, "concurrency", 0, "the refill rate for throttling. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().StringVar(&proxy, "proxy", "", "proxy for all theme requests. This will override what is in your config.yml")
	ThemeCmd.PersistentFlags().DurationVarP(&timeout, "timeout", "t", 0, "the timeout to kill any stalled processes. This will override what is in your config.yml")

	watchCmd.Flags().StringVarP(&notifyFile, "notify", "n", "", "file to touch when workers have gone idle")
	watchCmd.Flags().BoolVarP(&allenvs, "allenvs", "a", false, "run command with all environments")
	removeCmd.Flags().BoolVarP(&allenvs, "allenvs", "a", false, "run command with all environments")
	replaceCmd.Flags().BoolVarP(&allenvs, "allenvs", "a", false, "run command with all environments")
	uploadCmd.Flags().BoolVarP(&allenvs, "allenvs", "a", false, "run command with all environments")

	bootstrapCmd.Flags().StringVar(&bootstrapVersion, "version", latestRelease, "version of Shopify Timber to use")
	bootstrapCmd.Flags().StringVar(&bootstrapPrefix, "prefix", "", "prefix to the Timber theme being created")
	bootstrapCmd.Flags().BoolVar(&setThemeID, "setid", true, "update config with ID of created Theme")

	ThemeCmd.AddCommand(bootstrapCmd, removeCmd, replaceCmd, uploadCmd, watchCmd, downloadCmd, versionCmd, updateCmd, configureCmd)
}

func initializeConfig(cmdName string, timesout bool) error {
	if cmdName != "update" && isNewReleaseAvailable() {
		fmt.Println(kit.YellowText(fmt.Sprintf("%s\n%s\n%s", banner, updateAvailableMessage, banner)))
	}

	kit.SetFlagConfig(kit.Configuration{
		Password:    password,
		ThemeID:     themeid,
		Domain:      domain,
		Directory:   directory,
		Proxy:       proxy,
		BucketSize:  bucketsize,
		RefillRate:  refillrate,
		Concurrency: concurrency,
		Timeout:     timeout,
	})

	var err error
	if environments, err = kit.LoadEnvironments(configPath); err != nil {
		return err
	}

	eventLog := make(chan kit.ThemeEvent)
	themeClients = []kit.ThemeClient{}

	if allenvs {
		for env := range environments {
			config, err := environments.GetConfiguration(env)
			if err != nil {
				return err
			}
			themeClients = append(themeClients, kit.NewThemeClient(eventLog, config))
		}
	} else {
		config, err := environments.GetConfiguration(environment)
		if err != nil {
			return err
		}
		themeClients = []kit.ThemeClient{kit.NewThemeClient(eventLog, config)}
	}

	go consumeEventLog(eventLog, timesout, themeClients[0].GetConfiguration().Timeout)

	return nil
}

func consumeEventLog(eventLog chan kit.ThemeEvent, timesout bool, timeout time.Duration) {
	eventTicked := true
	for {
		select {
		case event := <-eventLog:
			eventTicked = true
			fmt.Printf("%s\n", event)
		case <-time.Tick(timeout):
			if !timesout {
				break
			}
			if !eventTicked {
				fmt.Printf("Theme Kit timed out after %v seconds\n", timeout)
				os.Exit(1)
			}
			eventTicked = false
		}
	}
}
