package util

import (
    "os"
    "github.com/op/go-logging"
)

func MustInitLogging(debugMode bool, withTime bool) {
    level := logging.INFO
    format := "%{level:.1s}: %{message}"
    timeFormat := ""

    if withTime {
        timeFormat = "%{time:2006.01.02 15:04:05} "
    }

    if debugMode {
        level = logging.DEBUG
        format = " %{shortfile} " + format
        timeFormat = "%{time:2006.01.02 15:04:05.000} "
    }

    format = timeFormat + format

    logging.SetBackend(logging.NewLogBackend(os.Stderr, "", 0))
    logging.SetFormatter(logging.MustStringFormatter(format))
    logging.SetLevel(level, "")
}

func MustGetLogger(name string) *logging.Logger {
    return logging.MustGetLogger(name)
}