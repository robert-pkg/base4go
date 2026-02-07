package main

import (
	"os"
	"time"

	"github.com/robert-pkg/base4go/log"
	zap_log "github.com/robert-pkg/base4go/log/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLog(logFileName string, enableConsoleOutput bool, fields map[string]interface{}) error {

	if len(logFileName) == 0 && !enableConsoleOutput {
		panic("no log config")
	}

	zCfg := zap.NewProductionConfig()

	zCfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	zCfg.Encoding = "json"
	zCfg.EncoderConfig.TimeKey = "t"
	zCfg.EncoderConfig.LevelKey = "l"
	zCfg.EncoderConfig.NameKey = "logger"
	zCfg.EncoderConfig.CallerKey = "c"
	zCfg.EncoderConfig.MessageKey = "msg"
	zCfg.EncoderConfig.StacktraceKey = "st"
	zCfg.EncoderConfig.LineEnding = zapcore.DefaultLineEnding
	zCfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	zCfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	zCfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	zCfg.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder //zapcore.ShortCallerEncoder

	coreList := make([]zapcore.Core, 0, 2)
	if enableConsoleOutput {
		consoleDebugging := zapcore.Lock(os.Stdout)
		consoleEncoder := zapcore.NewConsoleEncoder(zCfg.EncoderConfig)

		core := zapcore.NewCore(consoleEncoder, consoleDebugging, zCfg.Level)
		coreList = append(coreList, core)
	}

	if len(logFileName) > 0 {
		jsonEncoder := zapcore.NewJSONEncoder(zCfg.EncoderConfig)

		hook := &lumberjack.Logger{
			Filename:   logFileName, // 日志文件路径
			MaxSize:    100,         // 单个日志文件最大多少 mb
			MaxBackups: 10,          // 日志备份数量
			MaxAge:     10,          // 日志最长保留时间。 days
			LocalTime:  true,
			Compress:   true, // 是否压缩日志
		}

		fileWriter := zapcore.AddSync(hook)

		core := zapcore.NewCore(jsonEncoder, fileWriter, zCfg.Level)
		coreList = append(coreList, core)

	}

	logger := zap.New(zapcore.NewTee(coreList...),
		zap.AddCaller(), // 显示调用者
		zap.AddCallerSkip(2),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	l, err := zap_log.NewLogger(
		zap_log.WithConfig(zCfg),
		zap_log.WithLogger(logger),
		log.WithLevel(log.DebugLevel),
		log.WithCallerSkipCount(2),
		log.WithFields(fields),
	)
	if err != nil {
		return err
	}

	log.DefaultLogger = l
	return nil
}

func main() {
	err := InitLog("xxx.log", true, map[string]interface{}{"Server": "xxx"})
	if err != nil {
		panic(err)
	}

	log.Info("hello world")
	log.Info("hello world", "robert")
	log.Infof("hello world %s", "robert")
}
