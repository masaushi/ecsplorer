package ui

type defaultOptions struct {
	commands []string
}

func newDefaultOptions() *defaultOptions {
	// commandText := "▼ ▲ (j k): navigate, q: quit, esc: cancel, r: reload, ?: help"
	return &defaultOptions{
		commands: []string{"▼ ▲ (j k): navigate", "q: quit", "esc: cancel", "r: reload"},
	}
}

type Option func(*defaultOptions)

func WithAdditionalCommands(commands []string) Option {
	return func(opt *defaultOptions) {
		opt.commands = append(opt.commands, commands...)
	}
}
