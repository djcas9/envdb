package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"
)

type Level uint8

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

var (
	DebugPrefix = "[\033[32mDEBUG\033[0m]"
	InfoPrefix  = "[\033[34mINFO\033[0m]"
	WarnPrefix  = "[\033[33mWARN\033[0m]"
	ErrorPrefix = "[\033[31mERROR\033[0m]"
	FatalPrefix = "[\033[31mFATAL\033[0m]"
	PanicPrefix = "[\033[31mPANIC\033[0m]"
)

type Logger struct {
	Level      Level
	Out        io.Writer
	Prefix     string
	Time       bool
	TimeFormat string
}

func NewLogger() *Logger {
	return &Logger{
		Out:        os.Stdout,
		Level:      InfoLevel,
		Prefix:     "",
		Time:       false,
		TimeFormat: "15:04:05",
	}
}

func (self *Logger) SetLevel(l Level) {
	self.Level = l
}

func (self *Logger) write(format string) {
	data := bytes.NewBuffer([]byte(format))

	_, err := io.Copy(self.Out, data)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write to log, %v\n", err)
	}
}

func (self *Logger) output(prefix string, args ...interface{}) {
	self.write(fmt.Sprintf("%s %s %s %s\n",
		self.Prefix,
		time.Now().Format(self.TimeFormat),
		prefix,
		fmt.Sprint(args...),
	))
}

func (self *Logger) outputf(prefix, format string, args ...interface{}) {
	self.write(fmt.Sprintf("%s %s %s %s\n",
		self.Prefix,
		time.Now().Format(self.TimeFormat),
		prefix,
		fmt.Sprintf(format, args...),
	))
}

func (self *Logger) Debug(args ...interface{}) {
	if self.Level >= DebugLevel {
		self.output(DebugPrefix, args...)
	}
}

func (self *Logger) Debugf(format string, args ...interface{}) {
	if self.Level >= DebugLevel {
		self.outputf(DebugPrefix, format, args...)
	}
}

func (self *Logger) Info(args ...interface{}) {
	if self.Level >= InfoLevel {
		self.output(InfoPrefix, args...)
	}
}

func (self *Logger) Infof(format string, args ...interface{}) {
	if self.Level >= InfoLevel {
		self.outputf(InfoPrefix, format, args...)
	}
}

func (self *Logger) Warn(args ...interface{}) {
	if self.Level >= WarnLevel {
		self.output(WarnPrefix, args...)
	}
}

func (self *Logger) Warnf(format string, args ...interface{}) {
	if self.Level >= WarnLevel {
		self.outputf(WarnPrefix, format, args...)
	}
}

func (self *Logger) Error(args ...interface{}) {
	if self.Level >= ErrorLevel {
		self.output(ErrorPrefix, args...)
	}
}

func (self *Logger) Errorf(format string, args ...interface{}) {
	if self.Level >= ErrorLevel {
		self.outputf(ErrorPrefix, format, args...)
	}
}

func (self *Logger) Fatal(args ...interface{}) {
	if self.Level >= FatalLevel {
		self.output(FatalPrefix, args...)
		os.Exit(-1)
	}
}

func (self *Logger) Fatalf(format string, args ...interface{}) {
	if self.Level >= FatalLevel {
		self.outputf(FatalPrefix, format, args...)
		os.Exit(-1)
	}
}

func (self *Logger) Panic(args ...interface{}) {
	if self.Level >= PanicLevel {
		self.output(PanicPrefix, args...)
		os.Exit(-1)
	}
}

func (self *Logger) Panicf(format string, args ...interface{}) {
	if self.Level >= PanicLevel {
		self.outputf(PanicPrefix, format, args...)
		os.Exit(-1)
	}
}
