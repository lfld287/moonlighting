package base

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

type Logger struct {
	internal  *zap.Logger
	logrotate *lumberjack.Logger
}

func NewBaseLogger(
	path string,
	maxSize int,
	maxBackups int,
	maxAge int,
	compress bool,
	level zapcore.Level,
	stdout bool) *Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "line",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // UTC+8 时间格式\
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		//EncodeCaller:   zapcore.FullCallerEncoder,      // 全路径编码器
		EncodeName: zapcore.FullNameEncoder,
	}
	// 编码器
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	// 修改为添加lumberjack支持
	lj := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    maxSize, // 日志文件最大1M
		MaxBackups: maxBackups,
		MaxAge:     maxAge, // 日志保留最长时间7天
		Compress:   compress,
	}

	var writeSyncer zapcore.WriteSyncer
	if stdout {
		//打印到控制台和文件
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(lj))
	} else {
		//打印到文件
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(lj))
	}
	// 创建Logger
	core := zapcore.NewCore(encoder, writeSyncer, level)
	logger := zap.New(core, zap.AddCaller())

	return &Logger{
		internal:  logger,
		logrotate: lj,
	}
}

func (p *Logger) Debug(msg string, fields ...zap.Field) {
	p.internal.Debug(msg, fields...)
}

func (p *Logger) Info(msg string, fields ...zap.Field) {
	p.internal.Info(msg, fields...)
}

func (p *Logger) Warn(msg string, fields ...zap.Field) {
	p.internal.Warn(msg, fields...)
}

func (p *Logger) Error(msg string, fields ...zap.Field) {
	p.internal.Error(msg, fields...)
}

func (p *Logger) Sync() {
	_ = p.internal.Sync()
}
