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
	if !utils.Settings.GetBool("log-alert") {
		return
	}
	utils.Logger.Info("enable alert pusher")
	utils.Logger = utils.Logger.Named("go-httpguard-" + utils.Settings.GetString("host"))

	if utils.Settings.GetString("logger.push_api") != "" {
		// telegram alert
		alertPusher, err := utils.NewAlertPusherWithAlertType(
			ctx,
			utils.Settings.GetString("logger.push_api"),
			utils.Settings.GetString("logger.alert_type"),
			utils.Settings.GetString("logger.push_token"),
		)
		if err != nil {
			utils.Logger.Panic("create AlertPusher", zap.Error(err))
		}
		utils.Logger = utils.Logger.
			WithOptions(zap.HooksWithFields(alertPusher.GetZapHook()))
	}

	if utils.Settings.GetString("pateo_alert.push_api") != "" {
		// pateo wechat alert pusher
		pateoAlertPusher, err := utils.NewPateoAlertPusher(
			ctx,
			utils.Settings.GetString("pateo_alert.push_api"),
			utils.Settings.GetString("pateo_alert.token"),
		)
		if err != nil {
			utils.Logger.Panic("create PateoAlertPusher", zap.Error(err))
		}
		utils.Logger = utils.Logger.WithOptions(zap.HooksWithFields(pateoAlertPusher.GetZapHook()))
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
