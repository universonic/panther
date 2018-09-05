// Copyright 2018 Alfred Chou <unioverlord@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"regexp"

	zap "go.uber.org/zap"
	zapcore "go.uber.org/zap/zapcore"
)

const (
	// DEBUG logs are typically voluminous, and are usually disabled in production.
	DEBUG = zap.DebugLevel
	// INFO is the default logging priority.
	INFO = zap.InfoLevel
	// WARN logs are more important than Info, but don't need individual human
	// review.
	WARN = zap.WarnLevel
	// ERROR logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ERROR = zap.ErrorLevel
	// PANIC logs are particularly important errors. In development, the logger
	// panics after writing the message. In production, it logs message then
	// panics. The only difference is that it enables the possibility to output
	// more useful messages in development.
	PANIC = zap.DPanicLevel
	// FATAL logs a message, then calls os.Exit(1).
	FATAL = zap.FatalLevel
)

var (
	// DefaultLoggerFactory is the default LoggerFactory which only logs into
	// console. It is only used during program startup intialization.
	DefaultLoggerFactory, _ = NewFactoryConfig(
		[]string{"stdout"},
		[]string{"stderr"},
	).Apply(DEBUG, false)
	// DefaultSuperLogger is the underlay Logger
	DefaultSuperLogger = DefaultLoggerFactory.New()
	// DefaultLogger is the overlay SugaredLogger that is commonly used
	DefaultLogger = DefaultSuperLogger.Sugar()
)

// LoggerFactory is a factory to generate loggers. It is possible to use
// multiple loggers at the same time.
// TODO: join system logger inside
type LoggerFactory struct {
	cores []zapcore.Core
	opts  []zap.Option
}

// New exports merged cores with stored options as a Logger instance. For most
// cases, unless you have low-latency requirement, you should prevent using the
// Logger directly, but using SugaredLogger instead.
func (in *LoggerFactory) New() *zap.Logger {
	return zap.New(
		zapcore.NewTee(in.cores...),
		in.opts...,
	)
}

// WithOptions append options inside LoggerFactory to prepare for New()
func (in *LoggerFactory) WithOptions(opts ...zap.Option) *LoggerFactory {
	in.opts = append(in.opts, opts...)
	return in
}

// Merge combines multiple LoggerFactory into single one, with inherited cores
// and options.
// FIXME: options may have conflicts
func (in *LoggerFactory) Merge(factories ...*LoggerFactory) *LoggerFactory {
	for f := range factories {
		in.cores = append(in.cores, factories[f].cores...)
		in.WithOptions(factories[f].opts...)
	}
	return in
}

// NewLoggerFactory creates a non-option specified core-combined LoggerFactory.
// Commonly, you should create a FactoryConfig first, and then use Apply() to
// generate LoggerFactory, instead of using NewLoggerFactory directly.
// To append additional options, use WithOptions().
func NewLoggerFactory(cores ...zapcore.Core) *LoggerFactory {
	return &LoggerFactory{cores: cores}
}

// FactoryConfig indicates the path bundle used for generating LoggerFactory
type FactoryConfig struct {
	OutPath []string
	ErrPath []string
}

// Apply applys FactoryConfig to generate a new LoggerFactory.
// level specifys the lowest level that will be filed into log.
// highPriority is optional to specify the lowest level that should be filed
// into ErrPath, default is ERROR.
func (in *FactoryConfig) Apply(level zapcore.Level, jsoned bool, highPriority ...zapcore.Level) (*LoggerFactory, error) {
	hp := ERROR
	if len(highPriority) != 0 {
		hp = highPriority[0]
	}
	lowP := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < hp && lvl >= level
	})
	highP := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= hp && lvl >= level
	})
	if in.ErrPath == nil || len(in.ErrPath) == 0 {
		in.ErrPath = in.OutPath
	}
	lowSink, lowCloz, err := zap.Open(in.OutPath...)
	if err != nil {
		return nil, err
	}
	highSink, _, err := zap.Open(in.ErrPath...)
	if err != nil {
		lowCloz()
		return nil, err
	}
	levelEncoder := zapcore.CapitalLevelEncoder
	if len(in.OutPath) == 1 && regexp.MustCompile("stdout|stderr").Match([]byte(in.OutPath[0])) &&
		len(in.ErrPath) == 1 && regexp.MustCompile("stdout|stderr").Match([]byte(in.ErrPath[0])) {
		// Enable colored logging if we found this shall be a console logger
		levelEncoder = zapcore.CapitalColorLevelEncoder
	}
	encoderCnf := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "trace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    levelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	encoder := zapcore.NewConsoleEncoder(encoderCnf)
	if jsoned {
		encoder = zapcore.NewJSONEncoder(encoderCnf)
	}
	return NewLoggerFactory(
		zapcore.NewCore(encoder, lowSink, lowP),
		zapcore.NewCore(encoder, highSink, highP),
	), nil
}

// NewFactoryConfig generates a new FactoryConfig set with given path bundle
func NewFactoryConfig(outpath, errpath []string) *FactoryConfig {
	return &FactoryConfig{OutPath: outpath, ErrPath: errpath}
}
