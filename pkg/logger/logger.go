package logger

import (
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	log   *zap.Logger
	sugar *zap.SugaredLogger
}

type Config struct {
	Level    string
	Path     string
	Filename string
}

func New(debug bool) *Logger {
	cfg := Config{Level: "info", Path: "./logs", Filename: "app.log"}
	if debug {
		cfg.Level = "debug"
	}
	_ = os.MkdirAll(cfg.Path, 0755)
	lv := parseLevel(cfg.Level)

	fileWriter := &lumberjack.Logger{
		Filename:   filepath.Join(cfg.Path, cfg.Filename),
		MaxSize:    100,
		MaxBackups: 10,
		MaxAge:     30,
		Compress:   true,
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	consoleEnc := zapcore.NewConsoleEncoder(encoderConfig)
	fileEnc := zapcore.NewJSONEncoder(encoderConfig)

	core := zapcore.NewTee(
		zapcore.NewCore(fileEnc, zapcore.AddSync(io.Writer(fileWriter)), lv),
		zapcore.NewCore(consoleEnc, zapcore.AddSync(os.Stdout), lv),
	)
	z := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return &Logger{log: z, sugar: z.Sugar()}
}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func (l *Logger) Sync() error {
	if l == nil || l.log == nil {
		return nil
	}
	return l.log.Sync()
}

func (l *Logger) Debug(msg string, fields ...zap.Field) { l.log.Debug(msg, fields...) }
func (l *Logger) Info(msg string, fields ...zap.Field)  { l.log.Info(msg, fields...) }
func (l *Logger) Warn(msg string, fields ...zap.Field)  { l.log.Warn(msg, fields...) }
func (l *Logger) Error(msg string, fields ...zap.Field) { l.log.Error(msg, fields...) }
func (l *Logger) Fatal(msg string, fields ...zap.Field) { l.log.Fatal(msg, fields...) }

func (l *Logger) Debugf(tmpl string, args ...interface{}) { l.sugar.Debugf(tmpl, args...) }
func (l *Logger) Infof(tmpl string, args ...interface{})  { l.sugar.Infof(tmpl, args...) }
func (l *Logger) Warnf(tmpl string, args ...interface{})  { l.sugar.Warnf(tmpl, args...) }
func (l *Logger) Errorf(tmpl string, args ...interface{}) { l.sugar.Errorf(tmpl, args...) }
func (l *Logger) Fatalf(tmpl string, args ...interface{}) { l.sugar.Fatalf(tmpl, args...) }

var defaultLogger = New(true)

func Default() *Logger                 { return defaultLogger }
func Debug(msg string, fs ...zap.Field)  { defaultLogger.Debug(msg, fs...) }
func Info(msg string, fs ...zap.Field)   { defaultLogger.Info(msg, fs...) }
func Warn(msg string, fs ...zap.Field)   { defaultLogger.Warn(msg, fs...) }
func Error(msg string, fs ...zap.Field)  { defaultLogger.Error(msg, fs...) }
func Fatal(msg string, fs ...zap.Field)  { defaultLogger.Fatal(msg, fs...) }
func Debugf(t string, a ...interface{})  { defaultLogger.Debugf(t, a...) }
func Infof(t string, a ...interface{})   { defaultLogger.Infof(t, a...) }
func Warnf(t string, a ...interface{})   { defaultLogger.Warnf(t, a...) }
func Errorf(t string, a ...interface{})  { defaultLogger.Errorf(t, a...) }
func Fatalf(t string, a ...interface{})  { defaultLogger.Fatalf(t, a...) }
