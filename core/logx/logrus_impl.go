package logx

import (
	"fmt"
	"os"
	"time"

	"github.com/evalphobia/logrus_fluent"
	"github.com/sirupsen/logrus"
)

const (
	// for further detail of default values
	// please refer to https://github.com/fluent/fluent-logger-golang/blob/v1.4.0/fluent/fluent.go#L22
	defaultBufferLimit = 8 * 1024
	defaultMaxRetry    = 10
	defaultRetryWait   = 60000
	defaultTimeOut     = 3 * time.Second
)

type FluentConfig struct {
	Host        string
	Port        int
	Env         string
	MinLogLevel logrus.Level
	ServiceName string
}

func NewLogrusLogger() Logger {
	return logrusImpl{
		l: logrus.NewEntry(logrus.New()),
	}
}
func NewLogrusFluentLogger(c FluentConfig) Logger {
	l := logrus.New()
	l.SetLevel(c.MinLogLevel)
	newHook := func(levelTag string, levels ...logrus.Level) *logrus_fluent.FluentHook {
		hook, err := logrus_fluent.NewWithConfig(logrus_fluent.Config{
			Host:                c.Host,
			Port:                c.Port,
			AsyncConnect:        true,
			BufferLimit:         defaultBufferLimit,
			MaxRetry:            defaultMaxRetry,
			RetryWait:           defaultRetryWait,
			Timeout:             defaultTimeOut,
			DefaultMessageField: "message",
		})
		if err != nil {
			l.Fatalf("new log hook failed, err: %v", err)
		}
		host, _ := os.Hostname()
		hook.SetLevels(levels)
		hook.SetTag(fmt.Sprintf("%v.%v.%v.%v", c.ServiceName, c.Env, levelTag, host))
		hook.AddFilter("time", func(interface{}) interface{} {
			now := time.Now()
			return now.UTC().Format(time.RFC3339Nano)
		})
		return hook
	}
	l.AddHook(newHook("panic", logrus.PanicLevel))
	l.AddHook(newHook("fatal", logrus.FatalLevel))
	l.AddHook(newHook("error", logrus.ErrorLevel))
	l.AddHook(newHook("warn", logrus.WarnLevel))
	l.AddHook(newHook("info", logrus.InfoLevel))
	l.AddHook(newHook("debug", logrus.DebugLevel))
	logger := l.WithFields(map[string]interface{}{
		"service": c.ServiceName,
		"env":     c.Env,
	})
	return logrusImpl{
		l: logger,
	}
}

type logrusImpl struct {
	l *logrus.Entry
}

func (l logrusImpl) WithField(key string, value interface{}) Logger {
	l.l = l.l.WithField(key, value)
	return l
}
func (l logrusImpl) WithFields(fields map[string]interface{}) Logger {
	l.l = l.l.WithFields(fields)
	return l
}
func (l logrusImpl) WithError(err error) Logger {
	l.l = l.l.WithError(err)
	return l
}

func (l logrusImpl) Debugf(format string, args ...interface{}) {
	l.l.Debugf(format, args...)
}
func (l logrusImpl) Infof(format string, args ...interface{}) {
	l.l.Infof(format, args...)
}
func (l logrusImpl) Warnf(format string, args ...interface{}) {
	l.l.Warnf(format, args...)
}
func (l logrusImpl) Errorf(format string, args ...interface{}) {
	l.l.Errorf(format, args...)
}
func (l logrusImpl) Fatalf(format string, args ...interface{}) {
	l.l.Fatalf(format, args...)
}
func (l logrusImpl) Panicf(format string, args ...interface{}) {
	l.l.Panicf(format, args...)
}

func (l logrusImpl) Debug(args ...interface{}) {
	l.l.Debug(args...)
}
func (l logrusImpl) Info(args ...interface{}) {
	l.l.Info(args...)
}
func (l logrusImpl) Warn(args ...interface{}) {
	l.l.Warn(args...)
}
func (l logrusImpl) Error(args ...interface{}) {
	l.l.Error(args...)
}
func (l logrusImpl) Fatal(args ...interface{}) {
	l.l.Fatal(args...)
}
func (l logrusImpl) Panic(args ...interface{}) {
	l.l.Panic(args...)
}
func (l logrusImpl) Debugln(args ...interface{}) {
	l.l.Debugln(args...)
}
func (l logrusImpl) Infoln(args ...interface{}) {
	l.l.Infoln(args...)
}
func (l logrusImpl) Warnln(args ...interface{}) {
	l.l.Warnln(args...)
}
func (l logrusImpl) Errorln(args ...interface{}) {
	l.l.Errorln(args...)
}
func (l logrusImpl) Fatalln(args ...interface{}) {
	l.l.Fatalln(args...)
}
func (l logrusImpl) Panicln(args ...interface{}) {
	l.l.Panicln(args...)
}
