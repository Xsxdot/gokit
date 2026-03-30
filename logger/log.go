package logger

import (
	"context"
	"encoding/json"
	"gokit/consts"
	"sync"

	uuid "github.com/google/uuid"
	"github.com/openzipkin/zipkin-go"
	"github.com/sirupsen/logrus"
)

type Log struct {
	*logrus.Entry
}

var (
	log     *Log
	logOnce sync.Once
)

func InitLogger(level string) *Log {
	logOnce.Do(func() {
		logger := logrus.New()

		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})

		logLevel := logrus.InfoLevel
		switch level {
		case "debug":
			logLevel = logrus.DebugLevel
		case "warn":
			logLevel = logrus.WarnLevel
		case "info":
			logLevel = logrus.InfoLevel
		}
		logger.SetLevel(logLevel)

		log = &Log{Entry: logrus.NewEntry(logger)}
	})
	return log
}

func GetLogger() *Log {
	// 如果已经初始化，直接返回，全程无锁
	if log != nil {
		return log
	}

	logOnce.Do(func() {
		logger := logrus.New()
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
		logger.SetLevel(logrus.DebugLevel)
		log = &Log{Entry: logrus.NewEntry(logger)}
	})
	return log
}

func (l *Log) AddHook(hook logrus.Hook) {
	l.Entry.Logger.AddHook(hook)
}

func (l *Log) WithField(key string, value interface{}) *Log {
	return &Log{l.Entry.WithField(key, value)}
}

func (l *Log) GetLogger() *logrus.Entry {
	return l.Entry
}

func (l *Log) WithFields(arg interface{}) *Log {
	var jsonMap map[string]interface{}
	bytes, err := json.Marshal(arg)
	if err != nil {
		return l.WithField("arg", arg)
	}
	err = json.Unmarshal(bytes, &jsonMap)
	if err != nil {
		return l.WithField("arg", arg)
	}

	return &Log{l.Entry.WithFields(jsonMap)}
}

func (l *Log) WithEntryName(entryName string) *Log {
	return l.WithField("EntryName", entryName)
}

func (l *Log) WithErr(err error) *Log {
	if err == nil {
		return l
	}
	return l.WithField("Err", err.Error())
}

func (l *Log) WithTrace(ctx context.Context) *Log {
	var traceID string
	span := zipkin.SpanFromContext(ctx)
	if span == nil {
		var ok bool
		if traceID, ok = ctx.Value(consts.TraceKey).(string); !ok {
			traceID = uuid.New().String()
			ctx = context.WithValue(ctx, consts.TraceKey, traceID)
		}
	} else {
		traceID = span.Context().TraceID.String()
	}
	return l.WithField("TraceId", traceID)
}

func (l *Log) WithUserID(userId interface{}) *Log {
	return l.WithField("UserId", userId)
}

func (l *Log) WithOrderID(orderId interface{}) *Log {
	return l.WithField("OrderId", orderId)
}
