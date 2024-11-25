package logs

import (
	"github.com/op/go-logging"
	"os"
)

var Logger, _ = GetLogger("")

const defLevel = "INFO"

// InitLogger Receives the log level to be set in go-logging as a string. This method
// parses the string and set the level to the logger. If the level string is not
// valid an error is returned
func InitLogger(logLevel string) error {
	baseBackend := logging.NewLogBackend(os.Stdout, "", 0)
	format := logging.MustStringFormatter(
		"%{color}%{time:2006-01-02 15:04:05.000000-07:00} %{shortfile} %{shortfunc} ▶ %{level:.5s} %{color:reset} %{message}",
	)
	backendFormatter := logging.NewBackendFormatter(baseBackend, format)

	backendLeveled := logging.AddModuleLevel(backendFormatter)
	if logLevel == "" {
		logLevel = defLevel
	}

	logLevelCode, err := logging.LogLevel(logLevel)
	if err != nil {
		return err
	}
	backendLeveled.SetLevel(logLevelCode, "")

	logging.SetBackend(backendLeveled)
	return nil
}

func GetLogger(moduleName string) (*logging.Logger, error) {
	return logging.GetLogger(moduleName)
}
