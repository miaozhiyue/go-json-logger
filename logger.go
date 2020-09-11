package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Logger struct {
	Out          io.Writer
	Hooks        LevelHooks
	Formatter    Formatter
	ReportCaller bool
	Level        Level
	mu           MutexWrap
	entryPool    sync.Pool
	ExitFunc     exitFunc
}

type exitFunc func(int)

type MutexWrap struct {
	lock     sync.Mutex
	disabled bool
}

func (mw *MutexWrap) Lock() {
	if !mw.disabled {
		mw.lock.Lock()
	}
}

func (mw *MutexWrap) Unlock() {
	if !mw.disabled {
		mw.lock.Unlock()
	}
}

func (mw *MutexWrap) Disable() {
	mw.disabled = true
}

func New() *Logger {
	localWriter = os.Stderr
	return &Logger{
		Out:          os.Stderr,
		Formatter:    new(TextFormatter),
		Hooks:        make(LevelHooks),
		Level:        InfoLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}
}

func (logger *Logger) newEntry() *Entry {
	entry, ok := logger.entryPool.Get().(*Entry)
	if ok {
		return entry
	}
	return NewEntry(logger)
}

func (logger *Logger) releaseEntry(entry *Entry) {
	entry.Data = map[string]interface{}{}
	logger.entryPool.Put(entry)
}

func (logger *Logger) WithField(key string, value interface{}) *Entry {
	entry := logger.newEntry()
	defer logger.releaseEntry(entry)
	return entry.WithField(key, value)
}

func (logger *Logger) WithFields(fields Fields) *Entry {
	entry := logger.newEntry()
	defer logger.releaseEntry(entry)
	return entry.WithFields(fields)
}

func (logger *Logger) WithError(err error) *Entry {
	entry := logger.newEntry()
	defer logger.releaseEntry(entry)
	return entry.WithError(err)
}

func (logger *Logger) WithContext(ctx context.Context) *Entry {
	entry := logger.newEntry()
	defer logger.releaseEntry(entry)
	return entry.WithContext(ctx)
}

func (logger *Logger) WithTime(t time.Time) *Entry {
	entry := logger.newEntry()
	defer logger.releaseEntry(entry)
	return entry.WithTime(t)
}

func (logger *Logger) Logf(level Level, field interface{}, format string, args ...interface{}) {
	if logger.IsLevelEnabled(level) {
		entry := logger.newEntry()
		entry.Logf(level, field, format, args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Tracef(field interface{}, format string, args ...interface{}) {
	logger.Logf(TraceLevel, field, format, args...)
}

func (logger *Logger) Debugf(field interface{}, format string, args ...interface{}) {
	logger.Logf(DebugLevel, field, format, args...)
}

func (logger *Logger) Infof(field interface{}, format string, args ...interface{}) {
	logger.Logf(InfoLevel, field, format, args...)
}

func (logger *Logger) Printf(field interface{}, format string, args ...interface{}) {
	entry := logger.newEntry()
	entry.Printf(field, format, args...)
	logger.releaseEntry(entry)
}

func (logger *Logger) Warnf(field interface{}, format string, args ...interface{}) {
	logger.Logf(WarnLevel, field, format, args...)
}

func (logger *Logger) Warningf(field interface{}, format string, args ...interface{}) {
	logger.Warnf(field, format, args...)
}

func (logger *Logger) Errorf(field interface{}, format string, args ...interface{}) {
	logger.Logf(ErrorLevel, field, format, args...)
}

func (logger *Logger) Fatalf(field interface{}, format string, args ...interface{}) {
	logger.Logf(FatalLevel, field, format, args...)
	logger.Exit(1)
}

func (logger *Logger) Panicf(field interface{}, format string, args ...interface{}) {
	logger.Logf(PanicLevel, field, format, args...)
}

func (logger *Logger) Log(level Level, field interface{}, args ...interface{}) {
	if logger.IsLevelEnabled(level) {
		entry := logger.newEntry()
		entry.Log(level, field, args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Trace(field interface{}, args ...interface{}) {
	logger.Log(TraceLevel, field, args...)
}

func (logger *Logger) Debug(field interface{}, args ...interface{}) {
	logger.Log(DebugLevel, field, args...)
}

func (logger *Logger) Info(field interface{}, args ...interface{}) {
	logger.Log(InfoLevel, field, args...)
}

func (logger *Logger) Print(field interface{}, args ...interface{}) {
	entry := logger.newEntry()
	entry.Print(field, args...)
	logger.releaseEntry(entry)
}

func (logger *Logger) Warn(field interface{}, args ...interface{}) {
	logger.Log(WarnLevel, field, args...)
}

func (logger *Logger) Warning(field interface{}, args ...interface{}) {
	logger.Warn(field, args...)
}

func (logger *Logger) Error(field interface{}, args ...interface{}) {
	logger.Log(ErrorLevel, field, args...)
}

func (logger *Logger) Fatal(field interface{}, args ...interface{}) {
	logger.Log(FatalLevel, field, args...)
	logger.Exit(1)
}

func (logger *Logger) Panic(field interface{}, args ...interface{}) {
	logger.Log(PanicLevel, field, args...)
}

func (logger *Logger) Logln(level Level, field interface{}, args ...interface{}) {
	if logger.IsLevelEnabled(level) {
		entry := logger.newEntry()
		entry.Logln(level, field, args...)
		logger.releaseEntry(entry)
	}
}

func (logger *Logger) Traceln(field interface{}, args ...interface{}) {
	logger.Logln(TraceLevel, field, args...)
}

func (logger *Logger) Debugln(field interface{}, args ...interface{}) {
	logger.Logln(DebugLevel, field, args...)
}

func (logger *Logger) Infoln(field interface{}, args ...interface{}) {
	logger.Logln(InfoLevel, field, args...)
}

func (logger *Logger) Println(field interface{}, args ...interface{}) {
	entry := logger.newEntry()
	entry.Println(field, args...)
	logger.releaseEntry(entry)
}

func (logger *Logger) Warnln(field interface{}, args ...interface{}) {
	logger.Logln(WarnLevel, field, args...)
}

func (logger *Logger) Warningln(field interface{}, args ...interface{}) {
	logger.Warnln(field, args...)
}

func (logger *Logger) Errorln(field interface{}, args ...interface{}) {
	logger.Logln(ErrorLevel, field, args...)
}

func (logger *Logger) Fatalln(field interface{}, args ...interface{}) {
	logger.Logln(FatalLevel, field, args...)
	logger.Exit(1)
}

func (logger *Logger) Panicln(field interface{}, args ...interface{}) {
	logger.Logln(PanicLevel, field, args...)
}

func (logger *Logger) Exit(code int) {
	runHandlers()
	if logger.ExitFunc == nil {
		logger.ExitFunc = os.Exit
	}
	logger.ExitFunc(code)
}

func (logger *Logger) SetNoLock() {
	logger.mu.Disable()
}

func (logger *Logger) level() Level {
	return Level(atomic.LoadUint32((*uint32)(&logger.Level)))
}

func (logger *Logger) SetLevel(level Level) {
	atomic.StoreUint32((*uint32)(&logger.Level), uint32(level))
}

func (logger *Logger) GetLevel() Level {
	return logger.level()
}

func (logger *Logger) AddHook(hook Hook) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.Hooks.Add(hook)
}

func (logger *Logger) IsLevelEnabled(level Level) bool {
	return logger.level() >= level
}

func (logger *Logger) SetFormatter(formatter Formatter) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.Formatter = formatter
}

func (logger *Logger) SetOutput(output io.Writer) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	localWriter = output
	logger.Out = output
}

func (logger *Logger) SetReportCaller(reportCaller bool) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.ReportCaller = reportCaller
}

func (logger *Logger) ReplaceHooks(hooks LevelHooks) LevelHooks {
	logger.mu.Lock()
	oldHooks := logger.Hooks
	logger.Hooks = hooks
	logger.mu.Unlock()
	return oldHooks
}

type Fields map[string]interface{}

type Level uint32

func (level Level) String() string {
	if b, err := level.MarshalText(); err == nil {
		return string(b)
	} else {
		return "unknown"
	}
}

func ParseLevel(lvl string) (Level, error) {
	switch strings.ToLower(lvl) {
	case "panic":
		return PanicLevel, nil
	case "fatal":
		return FatalLevel, nil
	case "error":
		return ErrorLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	case "trace":
		return TraceLevel, nil
	}

	var l Level
	return l, fmt.Errorf("not a valid logger Level: %q", lvl)
}

func (level *Level) UnmarshalText(text []byte) error {
	l, err := ParseLevel(string(text))
	if err != nil {
		return err
	}

	*level = l

	return nil
}

func (level Level) MarshalText() ([]byte, error) {
	switch level {
	case TraceLevel:
		return []byte("trace"), nil
	case DebugLevel:
		return []byte("debug"), nil
	case InfoLevel:
		return []byte("info"), nil
	case WarnLevel:
		return []byte("warning"), nil
	case ErrorLevel:
		return []byte("error"), nil
	case FatalLevel:
		return []byte("fatal"), nil
	case PanicLevel:
		return []byte("panic"), nil
	}

	return nil, fmt.Errorf("not a valid logger level %d", level)
}

var AllLevels = []Level{
	PanicLevel,
	FatalLevel,
	ErrorLevel,
	WarnLevel,
	InfoLevel,
	DebugLevel,
	TraceLevel,
}

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

var (
	_           = &log.Logger{}
	_ StdLogger = &Entry{}
	_ StdLogger = &Logger{}
)

type StdLogger interface {
	Print(interface{}, ...interface{})
	Printf(interface{}, string, ...interface{})
	Println(interface{}, ...interface{})

	Fatal(interface{}, ...interface{})
	Fatalf(interface{}, string, ...interface{})
	Fatalln(interface{}, ...interface{})

	Panic(interface{}, ...interface{})
	Panicf(interface{}, string, ...interface{})
	Panicln(interface{}, ...interface{})
}
