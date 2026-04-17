package logx

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// newTestLogger builds a logrusImpl with a discarded output and a no-op ExitFunc so
// Fatal* methods do not terminate the test process.
func newTestLogger(t *testing.T) (logrusImpl, *bytes.Buffer) {
	t.Helper()
	base := logrus.New()
	buf := &bytes.Buffer{}
	base.Out = buf
	base.SetLevel(logrus.TraceLevel)
	base.ExitFunc = func(int) {}
	return logrusImpl{l: logrus.NewEntry(base)}, buf
}

func TestNewLogrusLogger(t *testing.T) {
	l := NewLogrusLogger()
	assert.NotNil(t, l)

	impl, ok := l.(logrusImpl)
	assert.True(t, ok)
	// prevent noise on stdout for any real calls
	impl.l.Logger.Out = ioutil.Discard
}

func TestLogrusImpl_WithField_WithFields_WithError(t *testing.T) {
	l, _ := newTestLogger(t)
	l2 := l.WithField("k", "v")
	assert.NotNil(t, l2)
	l3 := l.WithFields(map[string]interface{}{"a": 1, "b": 2})
	assert.NotNil(t, l3)
	l4 := l.WithError(assert.AnError)
	assert.NotNil(t, l4)
}

func TestLogrusImpl_FormatMethods(t *testing.T) {
	l, buf := newTestLogger(t)
	l.Debugf("debug %d", 1)
	l.Infof("info %d", 1)
	l.Warnf("warn %d", 1)
	l.Errorf("error %d", 1)
	l.Printf("print %d", 1)
	l.Tracef("trace %d", 1)
	l.Fatalf("fatal %d", 1)
	assert.NotPanics(t, func() {
		defer func() { _ = recover() }()
		l.Panicf("panic %d", 1)
	})
	assert.NotEmpty(t, buf.String())
}

func TestLogrusImpl_PlainMethods(t *testing.T) {
	l, buf := newTestLogger(t)
	l.Debug("a")
	l.Info("a")
	l.Warn("a")
	l.Error("a")
	l.Trace("a")
	l.Print("a")
	l.Fatal("a")
	assert.NotPanics(t, func() {
		defer func() { _ = recover() }()
		l.Panic("a")
	})
	assert.NotEmpty(t, buf.String())
}

func TestLogrusImpl_LnMethods(t *testing.T) {
	l, buf := newTestLogger(t)
	l.Debugln("a")
	l.Infoln("a")
	l.Warnln("a")
	l.Errorln("a")
	l.Traceln("a")
	l.Println("a")
	l.Fatalln("a")
	assert.NotPanics(t, func() {
		defer func() { _ = recover() }()
		l.Panicln("a")
	})
	assert.NotEmpty(t, buf.String())
}

func TestLogrusImpl_Logf(t *testing.T) {
	l, _ := newTestLogger(t)
	levels := []Level{PanicLevel, FatalLevel, ErrorLevel, WarnLevel, InfoLevel, DebugLevel, TraceLevel, Level(99)}
	for _, lvl := range levels {
		func() {
			defer func() { _ = recover() }()
			l.Logf(lvl, "msg %d", 1)
		}()
	}
}

func TestLogrusImpl_Log(t *testing.T) {
	l, _ := newTestLogger(t)
	levels := []Level{PanicLevel, FatalLevel, ErrorLevel, WarnLevel, InfoLevel, DebugLevel, TraceLevel, Level(99)}
	for _, lvl := range levels {
		func() {
			defer func() { _ = recover() }()
			l.Log(lvl, "m")
		}()
	}
}

func TestLogrusImpl_Logln(t *testing.T) {
	l, _ := newTestLogger(t)
	levels := []Level{PanicLevel, FatalLevel, ErrorLevel, WarnLevel, InfoLevel, DebugLevel, TraceLevel, Level(99)}
	for _, lvl := range levels {
		func() {
			defer func() { _ = recover() }()
			l.Logln(lvl, "m")
		}()
	}
}

func TestNewLogrusFluentLogger(t *testing.T) {
	cfg := FluentConfig{
		Host:        "127.0.0.1",
		Port:        0,
		Env:         "test",
		MinLogLevel: logrus.DebugLevel,
		ServiceName: "trinity-go-test",
	}
	assert.NotPanics(t, func() {
		logger := NewLogrusFluentLogger(cfg)
		assert.NotNil(t, logger)
		// Redirect the underlying logger output and exercise each level so the
		// registered hook filters are invoked (covers the time filter closure).
		impl := logger.(logrusImpl)
		impl.l.Logger.Out = ioutil.Discard
		impl.l.Logger.ExitFunc = func(int) {}
		impl.Debug("d")
		impl.Info("i")
		impl.WithField("time", "anything").Info("with-time")
		impl.Warn("w")
		impl.Error("e")
		func() {
			defer func() { _ = recover() }()
			impl.Panic("p")
		}()
		impl.Fatal("f")
	})
}
