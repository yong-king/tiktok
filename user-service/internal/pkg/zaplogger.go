package pkg

import (
	"io"
	"os"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapLogger 实现 Kratos log.Logger 接口，封装 zap.SugaredLogger
// 用于在 Kratos 中无侵入替换日志，支持控制台打印与文件落盘
// 通过 zapcore、lumberjack 实现按大小切割、保留历史、压缩归档
type zapLogger struct {
	log *zap.SugaredLogger
}

// Log 实现 Kratos log.Logger 接口的方法，将 Kratos Level 映射到 zap 方法
func (l *zapLogger) Log(level log.Level, keyvals ...interface{}) error {
	switch level {
	case log.LevelDebug:
		l.log.Debugw("", keyvals...)
	case log.LevelInfo:
		l.log.Infow("", keyvals...)
	case log.LevelWarn:
		l.log.Warnw("", keyvals...)
	case log.LevelError:
		l.log.Errorw("", keyvals...)
	default:
		l.log.Infow("", keyvals...)
	}
	return nil
}

type LogConfig struct {
	Level      string `yaml:"level"`
	Path       string `yaml:"path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
	Console    bool   `yaml:"console"`
}

// NewZapLogger 创建可直接用于 Kratos log.With 的 zapLogger 实例
// 落盘到 logs/app.log（JSON），控制台默认 io.Discard，如需打印替换为 os.Stdout
func NewZapLogger(cfg LogConfig) log.Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.Path,
		MaxSize:    cfg.MaxSize,    // 单文件最大 100MB
		MaxBackups: cfg.MaxBackups, // 最多保留 10 个历史文件
		MaxAge:     cfg.MaxAge,     // 文件最大保存天数
		Compress:   cfg.Compress,   // 是否压缩归档
	})

	var consoleWriter zapcore.WriteSyncer = zapcore.AddSync(io.Discard)
	if cfg.Console {
		consoleWriter = zapcore.AddSync(os.Stdout)
	}

	level := zap.InfoLevel
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn", "warning":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(fileWriter, consoleWriter),
		level,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
	return &zapLogger{log: logger}
}

var _ log.Logger = (*zapLogger)(nil)
