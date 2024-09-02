package xcodebuild

import "github.com/dashwave/go-utils/v2/command"

// CommandModel ...
type CommandModel interface {
	PrintableCmd() string
	Command(opts *command.Opts) command.Command
}
