package log

import (
	log "github.com/Sirupsen/logrus"
)

/*
const (
	DEBUG   = 1
	INFO    = 2
	WARNING = 4
	ERROR   = 8

	LOGOUT  = DEBUG | INFO | WARNING | ERROR
	LOGERR  = WARNING | ERROR
	OUTPUT  = INFO | WARNING | ERROR
	VERBOSE = ERROR
)
*/

func Debug(args ...interface{}) {
	log.Debug(args...)
}
func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Info(args ...interface{}) {
	log.Info(args...)
}
func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Warn(args ...interface{}) {
	log.Warn(args...)
}
func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}
func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}
