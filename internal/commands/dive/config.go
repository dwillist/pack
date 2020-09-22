package dive

import (
	"github.com/apex/log"
	"github.com/spf13/viper"
	"github.com/wagoodman/dive/dive/filetree"
	"strings"
)

func initConfig() {

	viper.SetDefault("log.level", log.InfoLevel.String())
	viper.SetDefault("log.path", "./dive.log")
	viper.SetDefault("log.enabled", false)
	// keybindings: status view / global
	viper.SetDefault("keybinding.quit", "ctrl+c")
	viper.SetDefault("keybinding.toggle-view", "tab")
	viper.SetDefault("keybinding.filter-files", "ctrl+f, ctrl+slash")
	// keybindings: layer view
	viper.SetDefault("keybinding.compare-all", "ctrl+a")
	viper.SetDefault("keybinding.compare-layer", "ctrl+l")
	// keybindings: filetree view
	viper.SetDefault("keybinding.toggle-collapse-dir", "space")
	viper.SetDefault("keybinding.toggle-collapse-all-dir", "ctrl+space")
	viper.SetDefault("keybinding.toggle-filetree-attributes", "ctrl+b")
	viper.SetDefault("keybinding.toggle-added-files", "ctrl+a")
	viper.SetDefault("keybinding.toggle-removed-files", "ctrl+r")
	viper.SetDefault("keybinding.toggle-modified-files", "ctrl+m")
	viper.SetDefault("keybinding.toggle-unmodified-files", "ctrl+u")
	viper.SetDefault("keybinding.page-up", "pgup")
	viper.SetDefault("keybinding.page-down", "pgdn")

	viper.SetDefault("diff.hide", "")

	viper.SetDefault("layer.show-aggregated-changes", false)

	viper.SetDefault("filetree.collapse-dir", false)
	viper.SetDefault("filetree.pane-width", 0.5)
	viper.SetDefault("filetree.show-attributes", true)

	viper.SetDefault("container-engine", "docker")
	viper.SetDefault("ignore-errors", false)

	//err = viper.BindPFlag("source", rootCmd.PersistentFlags().Lookup("source"))
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}

	viper.SetEnvPrefix("DIVE")
	// replace all - with _ when looking for matching environment variables
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// if config files are present, load them
	//if cfgFile == "" {
	//	// default configs are ignored if not found
	//	filepathToCfg := getDefaultCfgFile()
	//	viper.SetConfigFile(filepathToCfg)
	//} else {
	//	viper.SetConfigFile(cfgFile)
	//}
	//err = viper.ReadInConfig()
	//if err == nil {
	//	fmt.Println("Using config file:", viper.ConfigFileUsed())
	//} else if cfgFile != "" {
	//	fmt.Println(err)
	//	os.Exit(0)
	//}

	// set global defaults (for performance)
	filetree.GlobalFileTreeCollapse = viper.GetBool("filetree.collapse-dir")
}

