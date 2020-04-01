package proxy

import log "github.com/sirupsen/logrus"

type Logger struct {
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	if log.IsLevelEnabled(log.DebugLevel) {
		log.WithFields(l.getFields(args)).Debug(msg)
	}
}

func (l *Logger) Info(msg string, args ...interface{}) {
	if log.IsLevelEnabled(log.InfoLevel) {
		log.WithFields(l.getFields(args)).Info(msg)
	}
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	if log.IsLevelEnabled(log.WarnLevel) {
		log.WithFields(l.getFields(args)).Warn(msg)
	}
}

func (l *Logger) Error(msg string, args ...interface{}) {
	if log.IsLevelEnabled(log.ErrorLevel) {
		log.WithFields(l.getFields(args)).Error(msg)
	}
}

func (l *Logger) Panic(msg string, args ...interface{}) {
	log.WithFields(l.getFields(args)).Panic(msg)
}

func (l *Logger) getFields(args []interface{}) log.Fields {
	if len(args)%2 != 0 {
		panic("Length of args %!= 2")
	}

	fields := make(log.Fields, len(args)/2) // nolint: gomnd
	for i := 0; i < len(args); i += 2 {
		fields[args[i].(string)] = args[i+1]
	}

	return fields
}
