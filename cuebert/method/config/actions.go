package config

import (
	"github.com/johnmikee/cuebert/cuebert/method"
	"github.com/johnmikee/cuebert/cuebert/method/manager"
	"github.com/johnmikee/cuebert/cuebert/method/timebound"
)

type Config struct {
	Method method.Actions
}

type Method struct {
	Method method.Option
	Config method.Config
}

func New(m *Method) method.Actions {
	method := Config{
		Method: createMethod(m.Method),
	}
	method.Method.Setup(m.Config)

	return method.Method
}

func createMethod(m method.Option) method.Actions {
	switch m {
	case method.Manager:
		return &manager.Manager{}
	case method.TimeBound:
		return &timebound.TimeBound{}
	default:
		return nil
	}
}
