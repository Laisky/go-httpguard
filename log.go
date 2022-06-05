package httpguard

import gutils "github.com/Laisky/go-utils"

var Logger gutils.LoggerItf

func init() {
	var err error
	Logger, err = gutils.NewConsoleLoggerWithName("go-httpguard", gutils.LoggerLevelInfo)
	if err != nil {
		panic(err)
	}
}
