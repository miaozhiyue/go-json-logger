package logger

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	bufferPool *sync.Pool

	loggorPackage string

	minimumCallerDepth int

	callerInitOnce sync.Once
)

const (
	maximumCallerDepth int = 25
	knownLogrusFrames  int = 4
)

func init() {
	bufferPool = &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	minimumCallerDepth = 1
}

var ErrorKey = "error"

type Entry struct {
	Logger *Logger

	Data Fields

	Time time.Time

	Level Level

	Caller *runtime.Frame

	Message string

	Buffer *bytes.Buffer

	Context context.Context

	err string

	Field interface{}
}

func NewEntry(logger *Logger) *Entry {
	return &Entry{
		Logger: logger,
		Data:   make(Fields, 6),
	}
}

func (entry *Entry) Bytes() ([]byte, error) {
	chooseFile()
	return entry.Logger.Formatter.Format(entry)
}

func (entry *Entry) String() (string, error) {
	serialized, err := entry.Bytes()
	if err != nil {
		return "", err
	}
	str := string(serialized)
	return str, nil
}

func (entry *Entry) WithError(err error) *Entry {
	return entry.WithField(ErrorKey, err)
}

func (entry *Entry) WithContext(ctx context.Context) *Entry {
	dataCopy := make(Fields, len(entry.Data))
	for k, v := range entry.Data {
		dataCopy[k] = v
	}
	return &Entry{Logger: entry.Logger, Data: dataCopy, Time: entry.Time, err: entry.err, Context: ctx}
}

func (entry *Entry) WithField(key string, value interface{}) *Entry {
	return entry.WithFields(Fields{key: value})
}

func (entry *Entry) WithFields(fields Fields) *Entry {
	data := make(Fields, len(entry.Data)+len(fields))
	for k, v := range entry.Data {
		data[k] = v
	}
	fieldErr := entry.err
	for k, v := range fields {
		isErrField := false
		if t := reflect.TypeOf(v); t != nil {
			switch t.Kind() {
			case reflect.Func:
				isErrField = true
			case reflect.Ptr:
				isErrField = t.Elem().Kind() == reflect.Func
			}
		}
		if isErrField {
			tmp := fmt.Sprintf("can not add field %q", k)
			if fieldErr != "" {
				fieldErr = entry.err + ", " + tmp
			} else {
				fieldErr = tmp
			}
		} else {
			data[k] = v
		}
	}
	return &Entry{Logger: entry.Logger, Data: data, Time: entry.Time, err: fieldErr, Context: entry.Context}
}

func (entry *Entry) WithTime(t time.Time) *Entry {
	dataCopy := make(Fields, len(entry.Data))
	for k, v := range entry.Data {
		dataCopy[k] = v
	}
	return &Entry{Logger: entry.Logger, Data: dataCopy, Time: t, err: entry.err, Context: entry.Context}
}

func getPackageName(f string) string {
	for {
		lastPeriod := strings.LastIndex(f, ".")
		lastSlash := strings.LastIndex(f, "/")
		if lastPeriod > lastSlash {
			f = f[:lastPeriod]
		} else {
			break
		}
	}

	return f
}

func getCaller() *runtime.Frame {
	callerInitOnce.Do(func() {
		pcs := make([]uintptr, maximumCallerDepth)
		_ = runtime.Callers(0, pcs)
		for i := 0; i < maximumCallerDepth; i++ {
			funcName := runtime.FuncForPC(pcs[i]).Name()
			if strings.Contains(funcName, "getCaller") {
				loggorPackage = getPackageName(funcName)
				break
			}
		}
		minimumCallerDepth = knownLogrusFrames
	})

	pcs := make([]uintptr, maximumCallerDepth)
	depth := runtime.Callers(minimumCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	for f, again := frames.Next(); again; f, again = frames.Next() {
		pkg := getPackageName(f.Function)

		if pkg != loggorPackage {
			return &f
		}
	}
	return nil
}

func (entry Entry) HasCaller() (has bool) {
	return entry.Logger != nil &&
		entry.Logger.ReportCaller &&
		entry.Caller != nil
}

func (entry Entry) log(level Level, field interface{}, msg string) {
	var buffer *bytes.Buffer
	if entry.Time.IsZero() {
		entry.Time = time.Now()
	}
	entry.Level = level
	if len(msg) != 0 {
		entry.Message = msg
	}
	if nil != field && reflect.TypeOf(field).Kind() == reflect.Struct {
		entry.Field = field
	}
	entry.Logger.mu.Lock()
	if entry.Logger.ReportCaller {
		entry.Caller = getCaller()
	}
	entry.Logger.mu.Unlock()
	entry.fireHooks()
	buffer = bufferPool.Get().(*bytes.Buffer)
	buffer.Reset()
	defer bufferPool.Put(buffer)
	entry.Buffer = buffer
	entry.write()
	entry.Buffer = nil
	if level <= PanicLevel {
		panic(&entry)
	}
}

func (entry *Entry) fireHooks() {
	entry.Logger.mu.Lock()
	defer entry.Logger.mu.Unlock()
	err := entry.Logger.Hooks.Fire(entry.Level, entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to fire hook: %v\n", err)
	}
}

func (entry *Entry) write() {
	entry.Logger.mu.Lock()
	defer entry.Logger.mu.Unlock()
	chooseFile()
	serialized, err := entry.Logger.Formatter.Format(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to obtain reader, %v\n", err)
		return
	}
	if _, err = entry.Logger.Out.Write(serialized); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write to log, %v\n", err)
	}
}

func (entry *Entry) Log(level Level, field interface{}, args ...interface{}) {
	if entry.Logger.IsLevelEnabled(level) {
		entry.log(level, field, fmt.Sprint(args...))
	}
}

func (entry *Entry) Trace(field interface{}, args ...interface{}) {
	entry.Log(TraceLevel, field, args...)
}

func (entry *Entry) Debug(field interface{}, args ...interface{}) {
	entry.Log(DebugLevel, field, args...)
}

func (entry *Entry) Print(field interface{}, args ...interface{}) {
	entry.Info(field, args...)
}

func (entry *Entry) Info(field interface{}, args ...interface{}) {
	entry.Log(InfoLevel, field, args...)
}

func (entry *Entry) Warn(field interface{}, args ...interface{}) {
	entry.Log(WarnLevel, field, args...)
}

func (entry *Entry) Warning(field interface{}, args ...interface{}) {
	entry.Warn(field, args...)
}

func (entry *Entry) Error(field interface{}, args ...interface{}) {
	entry.Log(ErrorLevel, field, args...)
}

func (entry *Entry) Fatal(field interface{}, args ...interface{}) {
	entry.Log(FatalLevel, field, args...)
	entry.Logger.Exit(1)
}

func (entry *Entry) Panic(field interface{}, args ...interface{}) {
	entry.Log(PanicLevel, field, args...)
	panic(fmt.Sprint(args...))
}

func (entry *Entry) Logf(level Level, field interface{}, format string, args ...interface{}) {
	if entry.Logger.IsLevelEnabled(level) {
		entry.Log(level, field, fmt.Sprintf(format, args...))
	}
}

func (entry *Entry) Tracef(field interface{}, format string, args ...interface{}) {
	entry.Logf(TraceLevel, field, format, args...)
}

func (entry *Entry) Debugf(field interface{}, format string, args ...interface{}) {
	entry.Logf(DebugLevel, field, format, args...)
}

func (entry *Entry) Infof(field interface{}, format string, args ...interface{}) {
	entry.Logf(InfoLevel, field, format, args...)
}

func (entry *Entry) Printf(field interface{}, format string, args ...interface{}) {
	entry.Infof(field, format, args...)
}

func (entry *Entry) Warnf(field interface{}, format string, args ...interface{}) {
	entry.Logf(WarnLevel, field, format, args...)
}

func (entry *Entry) Warningf(field interface{}, format string, args ...interface{}) {
	entry.Warnf(field, format, args...)
}

func (entry *Entry) Errorf(field interface{}, format string, args ...interface{}) {
	entry.Logf(ErrorLevel, field, format, args...)
}

func (entry *Entry) Fatalf(field interface{}, format string, args ...interface{}) {
	entry.Logf(FatalLevel, field, format, args...)
	entry.Logger.Exit(1)
}

func (entry *Entry) Panicf(field interface{}, format string, args ...interface{}) {
	entry.Logf(PanicLevel, field, format, args...)
}

func (entry *Entry) Logln(level Level, field interface{}, args ...interface{}) {
	if entry.Logger.IsLevelEnabled(level) {
		entry.Log(level, field, entry.sprintlnn(args...))
	}
}

func (entry *Entry) Traceln(field interface{}, args ...interface{}) {
	entry.Logln(TraceLevel, field, args...)
}

func (entry *Entry) Debugln(field interface{}, args ...interface{}) {
	entry.Logln(DebugLevel, field, args...)
}

func (entry *Entry) Infoln(field interface{}, args ...interface{}) {
	entry.Logln(InfoLevel, field, args...)
}

func (entry *Entry) Println(field interface{}, args ...interface{}) {
	entry.Infoln(field, args...)
}

func (entry *Entry) Warnln(field interface{}, args ...interface{}) {
	entry.Logln(WarnLevel, field, args...)
}

func (entry *Entry) Warningln(field interface{}, args ...interface{}) {
	entry.Warnln(field, args...)
}

func (entry *Entry) Errorln(field interface{}, args ...interface{}) {
	entry.Logln(ErrorLevel, field, args...)
}

func (entry *Entry) Fatalln(field interface{}, args ...interface{}) {
	entry.Logln(FatalLevel, field, args...)
	entry.Logger.Exit(1)
}

func (entry *Entry) Panicln(field interface{}, args ...interface{}) {
	entry.Logln(PanicLevel, field, args...)
}

func (entry *Entry) sprintlnn(args ...interface{}) string {
	msg := fmt.Sprintln(args...)
	return msg[:len(msg)-1]
}
