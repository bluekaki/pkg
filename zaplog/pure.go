package zaplog

import (
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type writer struct {
	handler func(raw string)
}

func (w *writer) Write(raw []byte) (n int, err error) {
	w.handler(strings.TrimRight(string(raw), "\n"))
	return len(raw), nil
}

func (w *writer) Sync() error {
	return nil
}

// NewNonDurableJSONLogger return a json-encoder zap logger,
func NewNonDurableJSONLogger(handler func(raw string), opts ...Option) (*zap.Logger, error) {
	if handler == nil {
		panic("handler required")
	}

	opt := &option{level: DefaultLevel, fields: make(map[string]string)}
	for _, f := range opts {
		f(opt)
	}

	if opt.file != nil {
		panic("not-durable logger can't log into file")
	}

	timeLayout := DefaultTimeLayout
	if opt.timeLayout != "" {
		timeLayout = opt.timeLayout
	}

	// similar to zap.NewProductionEncoderConfig()
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "timestamp",
		LevelKey:      "level",
		NameKey:       "logger", // used by logger.Named(key); optional; useless
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace", // use by zap.AddStacktrace; optional; useless
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(timeLayout))
		},
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// lowPriority usd by info\debug\warn
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= opt.level && lvl < zapcore.ErrorLevel
	})

	// highPriority usd by error\panic\fatal
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= opt.level && lvl >= zapcore.ErrorLevel
	})

	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder,
			zapcore.AddSync(zapcore.Lock(&writer{handler: handler})),
			lowPriority,
		),
		zapcore.NewCore(jsonEncoder,
			zapcore.AddSync(zapcore.Lock(&writer{handler: handler})),
			highPriority,
		),
	)

	logger := zap.New(core,
		zap.AddCaller(),
		zap.ErrorOutput(zapcore.Lock(os.Stderr)),
	)

	for key, value := range opt.fields {
		logger = logger.WithOptions(zap.Fields(zapcore.Field{Key: key, Type: zapcore.StringType, String: value}))
	}
	return logger, nil
}
