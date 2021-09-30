package logs

import (
	"fmt"

	"github.com/JoaoLeal92/product-monitor-orchestrator/config"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	mongohook "github.com/weekface/mgorus"
)

type Logger struct {
	log       *logrus.Logger
	logFields map[string]interface{}
}

func NewLogger(cfg *config.LogConfig) *Logger {
	connectionString := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	hook, _ := mongohook.NewHooker(connectionString, cfg.Db, cfg.Collection)
	logFields := make(map[string]interface{})

	log := logrus.New()
	log.AddHook(hook)

	return &Logger{
		log:       log,
		logFields: logFields,
	}
}

func (l *Logger) AddFields(fields map[string]interface{}) {
	for key, value := range fields {
		l.logFields[key] = value
	}
}

func (l *Logger) ClearField(key string) {
	_, ok := l.logFields[key]
	if ok {
		delete(l.logFields, key)
	}
}

func (l *Logger) Info(message string) {
	additionalLogData := l.decodeArgsMap()

	l.log.WithFields(*additionalLogData).Info(message)
}

func (l *Logger) Warn(message string) {
	additionalLogData := l.decodeArgsMap()

	l.log.WithFields(*additionalLogData).Warn(message)
}

func (l *Logger) Error(message string) {
	additionalLogData := l.decodeArgsMap()

	l.log.WithFields(*additionalLogData).Error(message)
}

func (l *Logger) decodeArgsMap() *logrus.Fields {
	additionalLogData := logrus.Fields{}
	if len(l.logFields) > 0 {
		mapstructure.Decode(l.logFields, &additionalLogData)
	}

	return &additionalLogData
}
