package shared

import (
	"fmt"

	"github.com/backtesting-org/kronos-sdk/pkg/types/logging"
)

type logger struct {
}

func NewApplicationLogger() logging.ApplicationLogger {
	return &logger{}
}

func (l logger) Info(msg string, args ...interface{}) {
	fmt.Println(msg, args)
}

func (l logger) Debug(msg string, args ...interface{}) {
	fmt.Println(msg, args)

}

func (l logger) Warn(msg string, args ...interface{}) {
	fmt.Println(msg, args)

}

func (l logger) Error(msg string, args ...interface{}) {
	fmt.Println(msg, args)

}

func (l logger) Fatal(msg string, args ...interface{}) {
	fmt.Println(msg, args)

}

func (l logger) ErrorWithDebug(msg string, rawResponse []byte, args ...interface{}) {
	fmt.Println(msg, args)

}
