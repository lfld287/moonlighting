package console

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

type Logger struct {
	internal *zap.Logger
}

func NewConsoleLogger(level zapcore.Level) *Logger {
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

	var writeSyncer zapcore.WriteSyncer

	//打印到控制台和文件
	writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout))

	// 创建Logger
	core := zapcore.NewCore(encoder, writeSyncer, level)
	logger := zap.New(core, zap.AddCaller())

	return &Logger{
		internal: logger,
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
