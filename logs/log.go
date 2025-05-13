package logs

import (
	"context"
	"fmt"
	"github.com/drip-in/eden_lib/conf"
	"github.com/drip-in/eden_lib/el_utils/common_util"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

type Level = zapcore.Level

const (
	InfoLevel   Level = zap.InfoLevel   // 0, default level
	WarnLevel   Level = zap.WarnLevel   // 1
	ErrorLevel  Level = zap.ErrorLevel  // 2
	DPanicLevel Level = zap.DPanicLevel // 3, used in development log
	// PanicLevel logs a message, then panics
	PanicLevel Level = zap.PanicLevel // 4
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel Level = zap.FatalLevel // 5
	DebugLevel Level = zap.DebugLevel // -1
)

type Field = zap.Field

type LogFunc func(msg string, fields ...Field)
type LogCtxFunc func(ctx context.Context, msg string, fields ...Field)

func (l *Logger) Debug(msg string, fields ...Field) {
	if l.conf.ShowDebug {
		l.l.Debug(msg, fields...)
	}
}

func (l *Logger) CtxDebug(ctx context.Context, msg string, fields ...Field) {
	if l.conf.ShowDebug {
		l.l.Debug(fmt.Sprintf("%v  %v", LogIDFromContext(ctx), msg), fields...)
	}
}

func (l *Logger) Info(msg string, fields ...Field) {
	l.l.Info(msg, fields...)
}

func (l *Logger) CtxInfo(ctx context.Context, msg string, fields ...Field) {
	l.l.Info(fmt.Sprintf("%v  %v", LogIDFromContext(ctx), msg), fields...)
}

func (l *Logger) Warn(msg string, fields ...Field) {
	l.l.Warn(msg, fields...)
}

func (l *Logger) CtxWarn(ctx context.Context, msg string, fields ...Field) {
	l.l.Warn(fmt.Sprintf("%v  %v", LogIDFromContext(ctx), msg), fields...)
}

func (l *Logger) Error(msg string, fields ...Field) {
	l.l.Error(msg, fields...)
}

func (l *Logger) CtxError(ctx context.Context, msg string, fields ...Field) {
	l.l.Error(fmt.Sprintf("%v  %v", LogIDFromContext(ctx), msg), fields...)
}

func (l *Logger) DPanic(msg string, fields ...Field) {
	l.l.DPanic(msg, fields...)
}

func (l *Logger) CtxDPanic(ctx context.Context, msg string, fields ...Field) {
	l.l.DPanic(fmt.Sprintf("%v  %v", LogIDFromContext(ctx), msg), fields...)
}

func (l *Logger) Panic(msg string, fields ...Field) {
	l.l.Panic(msg, fields...)
}

func (l *Logger) CtxPanic(ctx context.Context, msg string, fields ...Field) {
	l.l.Panic(fmt.Sprintf("%v  %v", LogIDFromContext(ctx), msg), fields...)
}

func (l *Logger) Fatal(msg string, fields ...Field) {
	l.l.Fatal(msg, fields...)
}

func (l *Logger) CtxFatal(ctx context.Context, msg string, fields ...Field) {
	l.l.Fatal(fmt.Sprintf("%v  %v", LogIDFromContext(ctx), msg), fields...)
}

func (l *Logger) ConvertToFields(v ...interface{}) []Field {
	fields := make([]Field, 0)
	for _, arg := range v {
		if field, ok := arg.(Field); ok {
			fields = append(fields, field)
		}
	}
	return fields
}

// function variables for all field types
// in github.com/uber-go/zap/field.go

var (
	Skip        = zap.Skip
	Binary      = zap.Binary
	Bool        = zap.Bool
	Boolp       = zap.Boolp
	ByteString  = zap.ByteString
	Complex128  = zap.Complex128
	Complex128p = zap.Complex128p
	Complex64   = zap.Complex64
	Complex64p  = zap.Complex64p
	Float64     = zap.Float64
	Float64p    = zap.Float64p
	Float32     = zap.Float32
	Float32p    = zap.Float32p
	Int         = zap.Int
	Intp        = zap.Intp
	Int64       = zap.Int64
	Int64p      = zap.Int64p
	Int32       = zap.Int32
	Int32p      = zap.Int32p
	Int16       = zap.Int16
	Int16p      = zap.Int16p
	Int8        = zap.Int8
	Int8p       = zap.Int8p
	String      = zap.String
	Stringp     = zap.Stringp
	Uint        = zap.Uint
	Uintp       = zap.Uintp
	Uint64      = zap.Uint64
	Uint64p     = zap.Uint64p
	Uint32      = zap.Uint32
	Uint32p     = zap.Uint32p
	Uint16      = zap.Uint16
	Uint16p     = zap.Uint16p
	Uint8       = zap.Uint8
	Uint8p      = zap.Uint8p
	Uintptr     = zap.Uintptr
	Uintptrp    = zap.Uintptrp
	Reflect     = zap.Reflect
	Namespace   = zap.Namespace
	Stringer    = zap.Stringer
	Time        = zap.Time
	Timep       = zap.Timep
	Stack       = zap.Stack
	StackSkip   = zap.StackSkip
	Duration    = zap.Duration
	Durationp   = zap.Durationp
	Any         = zap.Any

	Info   LogFunc
	Infof  LogFunc
	Warn   LogFunc
	Error  LogFunc
	DPanic LogFunc
	Panic  LogFunc
	Fatal  LogFunc
	Debug  LogFunc

	CtxInfo   LogCtxFunc
	CtxInfof  LogCtxFunc
	CtxWarn   LogCtxFunc
	CtxError  LogCtxFunc
	CtxDPanic LogCtxFunc
	CtxPanic  LogCtxFunc
	CtxFatal  LogCtxFunc
	CtxDebug  LogCtxFunc
)

type Logger struct {
	l    *zap.Logger // zap ensure that zap.Logger is safe for concurrent use
	conf *conf.Zap
}

var std *Logger
var zapConf *conf.Zap

func (l *Logger) Sync() error {
	return l.l.Sync()
}

func Default() *Logger {
	return std
}

func Sync() error {
	if std != nil {
		return std.Sync()
	}
	return nil
}

func InitZap(conf *conf.Zap) {
	zapConf = conf
	if ok, _ := common_util.PathExists(zapConf.Director); !ok { // 判断是否有Director文件夹
		fmt.Printf("create %v directory\n", zapConf.Director)
		_ = os.Mkdir(zapConf.Director, os.ModePerm)
	}
	// 调试级别
	debugPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.DebugLevel
	})
	// 日志级别
	infoPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.InfoLevel
	})
	// 警告级别
	warnPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.WarnLevel
	})
	// 错误级别
	errorPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= zap.ErrorLevel
	})

	cores := [...]zapcore.Core{
		getEncoderCore(fmt.Sprintf("./%s/server_debug.log", zapConf.Director), debugPriority),
		getEncoderCore(fmt.Sprintf("./%s/server_info.log", zapConf.Director), infoPriority),
		getEncoderCore(fmt.Sprintf("./%s/server_warn.log", zapConf.Director), warnPriority),
		getEncoderCore(fmt.Sprintf("./%s/server_error.log", zapConf.Director), errorPriority),
	}
	logger := zap.New(zapcore.NewTee(cores[:]...))

	if zapConf.ShowLine {
		logger = logger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(1))
	}
	std = &Logger{
		l:    logger,
		conf: zapConf,
	}

	Info = std.Info
	Infof = std.Info
	Warn = std.Warn
	Error = std.Error
	DPanic = std.DPanic
	Panic = std.Panic
	Fatal = std.Fatal
	Debug = std.Debug

	CtxInfo = std.CtxInfo
	CtxInfof = std.CtxInfo
	CtxWarn = std.CtxWarn
	CtxError = std.CtxError
	CtxDPanic = std.CtxDPanic
	CtxPanic = std.CtxPanic
	CtxFatal = std.CtxFatal
	CtxDebug = std.CtxDebug
}

// getEncoderConfig 获取zapcore.EncoderConfig
func getEncoderConfig() (config zapcore.EncoderConfig) {
	config = zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  zapConf.StacktraceKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     CustomTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}
	switch zapConf.EncodeLevel {
	case "LowercaseLevelEncoder": // 小写编码器(默认)
		config.EncodeLevel = zapcore.LowercaseLevelEncoder
	case "LowercaseColorLevelEncoder": // 小写编码器带颜色
		config.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	case "CapitalLevelEncoder": // 大写编码器
		config.EncodeLevel = zapcore.CapitalLevelEncoder
	case "CapitalColorLevelEncoder": // 大写编码器带颜色
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	default:
		config.EncodeLevel = zapcore.LowercaseLevelEncoder
	}
	return config
}

// getEncoder 获取zapcore.Encoder
func getEncoder() zapcore.Encoder {
	if zapConf.Format == "json" {
		return zapcore.NewJSONEncoder(getEncoderConfig())
	}
	return zapcore.NewConsoleEncoder(getEncoderConfig())
}

// getEncoderCore 获取Encoder的zapcore.Core
func getEncoderCore(fileName string, level zapcore.LevelEnabler) (core zapcore.Core) {
	writer := getWriteSyncer(fileName) // 使用file-rotatelogs进行日志分割
	return zapcore.NewCore(getEncoder(), writer, level)
}

//@author: [SliverHorn](https://github.com/SliverHorn)
//@function: GetWriteSyncer
//@description: zap logger中加入file-rotatelogs
//@return: zapcore.WriteSyncer, error

func getWriteSyncer(file string) zapcore.WriteSyncer {
	if zapConf.LogInFile {
		lumberJackLogger := &lumberjack.Logger{
			Filename:   file, // 日志文件的位置
			MaxSize:    10,   // 在进行切割之前，日志文件的最大大小（以MB为单位）
			MaxBackups: 200,  // 保留旧文件的最大个数
			MaxAge:     30,   // 保留旧文件的最大天数
			Compress:   true, // 是否压缩/归档旧文件
		}
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(lumberJackLogger))
	}
	return zapcore.AddSync(os.Stdout)
}

// 自定义日志输出时间格式
func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(zapConf.Prefix + "2006/01/02 - 15:04:05.000"))
}
