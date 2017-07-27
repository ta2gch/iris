package runtime

import (
	"github.com/ta2gch/gazelle/runtime/class"
)

// Env struct is the struct for keeping functions and variables
type Env struct {
	Fun map[string]*class.Instance
	Var map[string]*class.Instance
}

// NewEnv creates new environment
func NewEnv() *Env {
	env := new(Env)
	env.Fun = map[string]*class.Instance{}
	env.Var = map[string]*class.Instance{}
	return env
}
