package main

import (
	"context"
	"fmt"

	gutils "github.com/Laisky/go-utils"
	gcmd "github.com/Laisky/go-utils/cmd"
	"github.com/Laisky/zap"
	"github.com/spf13/cobra"

	httpguard "github.com/Laisky/go-httpguard/v2"
)

var rootCMD = &cobra.Command{
	Use:   "go-httpguard",
	Short: "go-httpguard",
	Long:  `simple HTTP gateway`,
	Args:  gcmd.NoExtraArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		setupSettings()

		controller := httpguard.NewController(
			httpguard.NewAuth(
				httpguard.NewJwtAuthPlugin(httpguard.Config.JWTSecret),
				httpguard.NewBasicAuthPlugin(),
			),
			httpguard.NewAudit(),
			httpguard.NewBackend(),
		)
		err := controller.Run(ctx)
		if err != nil {
			panic(err)
		}
	},
}

func setupSettings() {
	var err error
	if err = gutils.Settings.LoadFromFile(httpguard.Config.Runtime.ConfigFile); err != nil {
		httpguard.Logger.Panic("load configuration",
			zap.Error(err),
			zap.String("config", httpguard.Config.Runtime.ConfigFile))
	}

	if err = gutils.Settings.Unmarshal(httpguard.Config); err != nil {
		httpguard.Logger.Panic("unmarshal config", zap.Error(err))
	}

	httpguard.Config.Init()

	fmt.Printf(">> \n%+v\n", httpguard.Config)
}

func init() {
	rootCMD.PersistentFlags().Bool("debug", false, "run in debug mode")
	rootCMD.PersistentFlags().StringVarP(&httpguard.Config.Runtime.ConfigFile,
		"config", "c",
		"/etc/go-httpguard/config.yml",
		"config file directory path")
}

func main() {
	if err := rootCMD.Execute(); err != nil {
		httpguard.Logger.Panic("start", zap.Error(err))
	}
}
