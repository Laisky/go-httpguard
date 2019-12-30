package main

import (
	"context"
	"fmt"

	"github.com/Laisky/zap"

	httpguard "github.com/Laisky/go-httpguard"
	"github.com/Laisky/go-utils"
	"github.com/spf13/pflag"
)

func setupSettings() {
	var err error
	if err = utils.Settings.SetupFromFile(utils.Settings.GetString("config")); err != nil {
		utils.Logger.Panic("load configuration", zap.Error(err), zap.String("config", utils.Settings.GetString("config")))
	}
}

func setupCommandArgs() {
	fmt.Println("start main...")
	pflag.Bool("debug", false, "run in debug mode")
	pflag.Bool("dry", false, "run in dry mode")
	pflag.String("config", "/etc/go-httpguard", "config file directory path")
	pflag.String("host", "", "hostname")
	pflag.String("log-level", "info", "log level")
	pflag.Parse()
	if err := utils.Settings.BindPFlags(pflag.CommandLine); err != nil {
		utils.Logger.Panic("parse arguments", zap.Error(err))
	}
}

func setupLogger(ctx context.Context) {
	var err error
	if utils.Settings.GetBool("debug") { // debug mode
		fmt.Println("run in debug mode")
		utils.Settings.Set("log-level", "debug")
		if err = utils.Logger.ChangeLevel("debug"); err != nil {
			utils.Logger.Panic("change logger level", zap.Error(err))
		}
	} else { // prod mode
		fmt.Println("run in prod mode")
		if err = utils.Logger.ChangeLevel("info"); err != nil {
			utils.Logger.Panic("change logger level", zap.Error(err))
		}
	}

	alertPusher, err := utils.NewAlertPusherWithAlertType(
		ctx,
		utils.Settings.GetString("logger.push_api"),
		utils.Settings.GetString("logger.alert_type"),
		utils.Settings.GetString("logger.push_token"),
	)
	if err != nil {
		utils.Logger.Panic("create AlertPusher", zap.Error(err))
	}

	hook := utils.NewAlertHook(alertPusher)
	if _, err := utils.SetDefaultLogger(
		"go-httpguard:"+utils.Settings.GetString("host"),
		utils.Settings.GetString("log-level"),
		zap.HooksWithFields(hook.GetZapHook())); err != nil {
		utils.Logger.Panic("setup logger", zap.Error(err))
	}
}

func main() {
	ctx := context.Background()
	setupCommandArgs()
	setupSettings()
	setupLogger(ctx)

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
