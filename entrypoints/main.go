package main

import (
	"fmt"

	httpguard "github.com/Laisky/go-httpguard"
	"github.com/Laisky/go-utils"
	"github.com/spf13/pflag"
)

func setupSettings() {
	utils.Settings.Setup(utils.Settings.GetString("config"))

	if utils.Settings.GetBool("debug") { // debug mode
		fmt.Println("run in debug mode")
		utils.SetupLogger("debug")
	} else { // prod mode
		fmt.Println("run in prod mode")
		utils.SetupLogger("info")
	}
}

func setupCommandArgs() {
	defer fmt.Println("All done")
	defer utils.Logger.Sync()
	fmt.Println("start main...")
	pflag.Bool("debug", false, "run in debug mode")
	pflag.Bool("dry", false, "run in dry mode")
	pflag.String("config", "/etc/go-httpguard", "config file directory path")
	pflag.Parse()
	utils.Settings.BindPFlags(pflag.CommandLine)
}

func main() {
	setupCommandArgs()
	setupSettings()

	controller := httpguard.NewController(
		httpguard.NewAuth(utils.Settings.GetString("secret")),
		httpguard.NewAudit(),
		httpguard.NewBackend(),
	)
	err := controller.Run()
	if err != nil {
		panic(err)
	}
}
