package container

import (
	"io/ioutil"

	"github.com/codeduckcloud/trinity-go/core/logx"
	"github.com/sirupsen/logrus"
)

func newTestLogger() logx.Logger {
	base := logrus.New()
	base.Out = ioutil.Discard
	base.SetLevel(logrus.PanicLevel)
	// prevent Fatal* from terminating the test process
	base.ExitFunc = func(int) {}
	return logrusWrapper{entry: logrus.NewEntry(base)}
}

// logrusWrapper exposes the same interface as logx.Logger but uses a
// locally-controlled logrus logger where ExitFunc is a no-op.
type logrusWrapper struct {
	entry *logrus.Entry
}

func (l logrusWrapper) WithField(key string, value interface{}) logx.Logger {
	return logrusWrapper{entry: l.entry.WithField(key, value)}
}
func (l logrusWrapper) WithFields(fields map[string]interface{}) logx.Logger {
	return logrusWrapper{entry: l.entry.WithFields(fields)}
}

func (l logrusWrapper) Debugf(format string, args ...interface{}) { l.entry.Debugf(format, args...) }
func (l logrusWrapper) Infof(format string, args ...interface{})  { l.entry.Infof(format, args...) }
func (l logrusWrapper) Warnf(format string, args ...interface{})  { l.entry.Warnf(format, args...) }
func (l logrusWrapper) Errorf(format string, args ...interface{}) { l.entry.Errorf(format, args...) }
func (l logrusWrapper) Fatalf(format string, args ...interface{}) { l.entry.Fatalf(format, args...) }
func (l logrusWrapper) Panicf(format string, args ...interface{}) { l.entry.Panicf(format, args...) }
func (l logrusWrapper) Printf(format string, args ...interface{}) { l.entry.Printf(format, args...) }
func (l logrusWrapper) Tracef(format string, args ...interface{}) { l.entry.Tracef(format, args...) }
func (l logrusWrapper) Logf(lvl logx.Level, format string, args ...interface{}) {
	l.entry.Logf(logrus.Level(lvl), format, args...)
}

func (l logrusWrapper) Debug(args ...interface{}) { l.entry.Debug(args...) }
func (l logrusWrapper) Info(args ...interface{})  { l.entry.Info(args...) }
func (l logrusWrapper) Warn(args ...interface{})  { l.entry.Warn(args...) }
func (l logrusWrapper) Error(args ...interface{}) { l.entry.Error(args...) }
func (l logrusWrapper) Fatal(args ...interface{}) { l.entry.Fatal(args...) }
func (l logrusWrapper) Panic(args ...interface{}) { l.entry.Panic(args...) }
func (l logrusWrapper) Print(args ...interface{}) { l.entry.Print(args...) }
func (l logrusWrapper) Trace(args ...interface{}) { l.entry.Trace(args...) }
func (l logrusWrapper) Log(lvl logx.Level, args ...interface{}) {
	l.entry.Log(logrus.Level(lvl), args...)
}

func (l logrusWrapper) Debugln(args ...interface{}) { l.entry.Debugln(args...) }
func (l logrusWrapper) Infoln(args ...interface{})  { l.entry.Infoln(args...) }
func (l logrusWrapper) Warnln(args ...interface{})  { l.entry.Warnln(args...) }
func (l logrusWrapper) Errorln(args ...interface{}) { l.entry.Errorln(args...) }
func (l logrusWrapper) Fatalln(args ...interface{}) { l.entry.Fatalln(args...) }
func (l logrusWrapper) Panicln(args ...interface{}) { l.entry.Panicln(args...) }
func (l logrusWrapper) Println(args ...interface{}) { l.entry.Println(args...) }
func (l logrusWrapper) Traceln(args ...interface{}) { l.entry.Traceln(args...) }
func (l logrusWrapper) Logln(lvl logx.Level, args ...interface{}) {
	l.entry.Logln(logrus.Level(lvl), args...)
}

var (
	logger     = newTestLogger()
	logWithCtx = logx.NewCtx(logger)
)
