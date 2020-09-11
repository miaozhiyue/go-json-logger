package logger

import (
	"context"
	"io"
	"os"
	"strings"
	"time"
)

var (
	std           = New()
	timeFlag      string
	file          *os.File
	localWriter   io.Writer
	nameStr       string
	timeStr       string
	pathStr       string
	chooseFileFlg bool
)

func StandardLogger() *Logger {
	return std
}

func SetOutput(out io.Writer) {
	localWriter = out
	std.SetOutput(out)
}

func SetOutputFile(path string, nameFormatter string, timeFormatter string) {
	timeFlag = time.Now().Format(timeFormatter)
	file, err := os.OpenFile(path+strings.ReplaceAll(nameFormatter, "${time}", timeFlag), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if nil != err {
		panic(err)
	}
	if nil != std.Out {
		if nil != file {
			std.SetOutput(io.MultiWriter([]io.Writer{file, localWriter}...))
		} else {
			std.SetOutput(io.MultiWriter([]io.Writer{file, os.Stdout}...))
		}
	} else {
		std.SetOutput(file)
	}
	pathStr = path
	nameStr = nameFormatter
	timeStr = timeFormatter
	chooseFileFlg = true
}

func chooseFile() {
	if chooseFileFlg {
		timeF := time.Now().Format(timeStr)
		if file != nil && timeF != timeFlag {
			SetOutputFile(pathStr, nameStr, timeStr)
		}
	}
}

func SetFormatter(formatter Formatter) {
	std.SetFormatter(formatter)
}

func SetReportCaller(include bool) {
	std.SetReportCaller(include)
}

func SetLevel(level Level) {
	std.SetLevel(level)
}

func GetLevel() Level {
	return std.GetLevel()
}

func IsLevelEnabled(level Level) bool {
	return std.IsLevelEnabled(level)
}

func AddHook(hook Hook) {
	std.AddHook(hook)
}

func WithError(err error) *Entry {
	return std.WithField(ErrorKey, err)
}

func WithContext(ctx context.Context) *Entry {
	return std.WithContext(ctx)
}

func WithField(key string, value interface{}) *Entry {
	return std.WithField(key, value)
}

func WithFields(fields Fields) *Entry {
	return std.WithFields(fields)
}

func WithTime(t time.Time) *Entry {
	return std.WithTime(t)
}

func Trace(field interface{}, args ...interface{}) {
	std.Trace(field, args...)
}

func Debug(field interface{}, args ...interface{}) {
	std.Debug(field, args...)
}

func Print(field interface{}, args ...interface{}) {
	std.Print(field, args...)
}

func Info(field interface{}, args ...interface{}) {
	std.Info(field, args...)
}

func Warn(field interface{}, args ...interface{}) {
	std.Warn(field, args...)
}

func Warning(field interface{}, args ...interface{}) {
	std.Warning(field, args...)
}

func Error(field interface{}, args ...interface{}) {
	std.Error(field, args...)
}

func Panic(field interface{}, args ...interface{}) {
	std.Panic(field, args...)
}

func Fatal(field interface{}, args ...interface{}) {
	std.Fatal(field, args...)
}

func Tracef(field interface{}, format string, args ...interface{}) {
	std.Tracef(field, format, args...)
}

func Debugf(field interface{}, format string, args ...interface{}) {
	std.Debugf(field, format, args...)
}

func Printf(field interface{}, format string, args ...interface{}) {
	std.Printf(field, format, args...)
}

func Infof(field interface{}, format string, args ...interface{}) {
	std.Infof(field, format, args...)
}

func Warnf(field interface{}, format string, args ...interface{}) {
	std.Warnf(field, format, args...)
}

func Warningf(field interface{}, format string, args ...interface{}) {
	std.Warningf(field, format, args...)
}

func Errorf(field interface{}, format string, args ...interface{}) {
	std.Errorf(field, format, args...)
}

func Panicf(field interface{}, format string, args ...interface{}) {
	std.Panicf(field, format, args...)
}

func Fatalf(field interface{}, format string, args ...interface{}) {
	std.Fatalf(field, format, args...)
}

func Traceln(field interface{}, args ...interface{}) {
	std.Traceln(field, args...)
}

func Debugln(field interface{}, args ...interface{}) {
	std.Debugln(field, args...)
}

func Println(field interface{}, args ...interface{}) {
	std.Println(field, args...)
}

func Infoln(field interface{}, args ...interface{}) {
	std.Infoln(field, args...)
}

func Warnln(field interface{}, args ...interface{}) {
	std.Warnln(field, args...)
}

func Warningln(field interface{}, args ...interface{}) {
	std.Warningln(field, args...)
}

func Errorln(field interface{}, args ...interface{}) {
	std.Errorln(field, args...)
}

func Panicln(field interface{}, args ...interface{}) {
	std.Panicln(field, args...)
}

func Fatalln(field interface{}, args ...interface{}) {
	std.Fatalln(field, args...)
}
