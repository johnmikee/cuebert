package bot

import (
	"github.com/johnmikee/cuebert/bot/lifecycle"
	"github.com/johnmikee/cuebert/cuebert/config"
)

// Option is a type for the different options that can be registered
type Option func(*Bot)

// Update is a function that updates the bot with the given options
func (b *Bot) Update(opt ...Option) {
	for _, o := range opt {
		o(b)
	}
}

// WithFlags sets the flags for the bot
func WithFlags(flags *config.Flags) Option {
	return func(b *Bot) {
		b.flags = flags
	}
}

// WithInteractive sets the interactive interface for the bot
func WithInteractive(interactive Interactive) Option {
	return func(b *Bot) {
		b.interactive = interactive
	}
}

// WithLifeCycle sets the lifecycle interface for the bot
func WithLifeCycle(lifecycle lifecycle.LifeCycle) Option {
	return func(b *Bot) {
		b.lifecycle = lifecycle
	}
}
