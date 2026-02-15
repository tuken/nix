package logger

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	gormlog "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type Logger interface {
	gormlog.Interface
	Debugf(template string, args ...interface{})
	Debugw(msg string, keysAndValues ...interface{})
	Infof(template string, args ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnf(template string, args ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorf(template string, args ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatalf(template string, args ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
	With(args ...interface{}) Logger
}

func NewLogger(filename string) Logger {

	return &AppLogger{
		logger:        newZapLogger(filename),
		logLevel:      gormlog.Info,
		slowThreshold: time.Second,
	}
}

func newZapLogger(filename string) *zap.SugaredLogger {

	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(config)

	logFile, _ := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	writer := zapcore.AddSync(logFile)
	defaultLogLevel := zapcore.DebugLevel
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, defaultLogLevel),
	)

	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel)).Sugar()
}

type AppLogger struct {
	logger        *zap.SugaredLogger
	logLevel      gormlog.LogLevel
	slowThreshold time.Duration
}

func (l AppLogger) With(args ...interface{}) Logger {

	return &AppLogger{
		logger:        l.logger.With(args...),
		logLevel:      l.logLevel,
		slowThreshold: l.slowThreshold,
	}
}

func (l AppLogger) Debugf(template string, args ...interface{}) {

	l.logger.Debugf(template, args...)
}

func (l AppLogger) Debugw(msg string, keysAndValues ...interface{}) {

	l.logger.Debugw(msg, keysAndValues...)
}

func (l AppLogger) Infof(template string, args ...interface{}) {

	l.logger.Infof(template, args...)
}

func (l AppLogger) Infow(msg string, keysAndValues ...interface{}) {

	l.logger.Infow(msg, keysAndValues...)
}

func (l AppLogger) Warnf(template string, args ...interface{}) {

	l.logger.Warnf(template, args...)
}

func (l AppLogger) Warnw(msg string, keysAndValues ...interface{}) {

	l.logger.Warnw(msg, keysAndValues...)
}

func (l AppLogger) Errorf(template string, args ...interface{}) {

	l.logger.Errorf(template, args...)
}

func (l AppLogger) Errorw(msg string, keysAndValues ...interface{}) {

	l.logger.Errorw(msg, keysAndValues...)
}

func (l AppLogger) Fatalf(template string, args ...interface{}) {

	l.logger.Fatalf(template, args...)
}

func (l AppLogger) Fatalw(msg string, keysAndValues ...interface{}) {

	l.logger.Fatalw(msg, keysAndValues...)
}

func (l AppLogger) LogMode(level gormlog.LogLevel) gormlog.Interface {

	newLogger := l
	newLogger.logLevel = level
	return &newLogger
}

func (l AppLogger) Info(_ context.Context, msg string, data ...interface{}) {

	if l.logLevel >= gormlog.Info {
		l.logger.Infof(msg, data...)
	}
}

func (l AppLogger) Warn(_ context.Context, msg string, data ...interface{}) {

	if l.logLevel >= gormlog.Warn {
		l.logger.Warnf(msg, data...)
	}
}

func (l AppLogger) Error(_ context.Context, msg string, data ...interface{}) {

	if l.logLevel >= gormlog.Error {
		l.logger.Errorf(msg, data...)
	}
}

func (l AppLogger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {

	if l.logLevel >= 0 {

		elapsed := time.Since(begin)
		switch {
		case err != nil && l.logLevel >= gormlog.Error:
			sql, rows := fc()
			l.logger.Infof("%s %s\n[%.3fms] [rows:%d] %s", utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		case elapsed > l.slowThreshold && l.slowThreshold != 0 && l.logLevel >= gormlog.Warn:
			sql, rows := fc()
			l.logger.Infof("%s\n[%.3fms] [rows:%d] %s", utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		case l.logLevel >= gormlog.Info:
			sql, rows := fc()
			l.logger.Infof("%s\n[%.3fms] [rows:%d] %s", utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

func (l AppLogger) Write(p []byte) (n int, err error) {

	l.logger.Info(string(p))
	return len(p), nil
}
