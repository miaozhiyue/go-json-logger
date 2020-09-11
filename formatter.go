package logger

import "time"

const (
	defaultTimestampFormat = time.RFC3339
	FieldKeyMsg            = "message"
	FieldKeyLevel          = "level"
	FieldKeyTime           = "@timestamp"
	FieldKeyLoggorError    = "error"
	FieldKeyFunc           = "func"
	FieldKeyFile           = "file"
)

type Formatter interface {
	Format(*Entry) ([]byte, error)
}

func prefixFieldClashes(data Fields, fieldMap FieldMap, reportCaller bool) {
	timeKey := fieldMap.resolve(FieldKeyTime)
	if t, ok := data[timeKey]; ok {
		data["fields."+timeKey] = t
		delete(data, timeKey)
	}

	msgKey := fieldMap.resolve(FieldKeyMsg)
	if m, ok := data[msgKey]; ok {
		data["fields."+msgKey] = m
		delete(data, msgKey)
	}

	levelKey := fieldMap.resolve(FieldKeyLevel)
	if l, ok := data[levelKey]; ok {
		data["fields."+levelKey] = l
		delete(data, levelKey)
	}

	loggerErrKey := fieldMap.resolve(FieldKeyLoggorError)
	if l, ok := data[loggerErrKey]; ok {
		data["fields."+loggerErrKey] = l
		delete(data, loggerErrKey)
	}

	if reportCaller {
		funcKey := fieldMap.resolve(FieldKeyFunc)
		if l, ok := data[funcKey]; ok {
			data["fields."+funcKey] = l
		}
		fileKey := fieldMap.resolve(FieldKeyFile)
		if l, ok := data[fileKey]; ok {
			data["fields."+fileKey] = l
		}
	}
}
