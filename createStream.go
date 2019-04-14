package reka

import (
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
)

func CreateStream() *Stream {
	logger := &log.Logger{
		Handler: cli.New(os.Stdout),
	}

	return &Stream{chains: &tree{}, Logger: logger}
}
