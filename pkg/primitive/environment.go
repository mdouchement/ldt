package primitive

import "os"

// An Env is used to export a new env and restore the original.
type Env struct {
	original map[string]string
	new      map[string]string
}

// NewEnv returns a new Env.
func NewEnv(env map[string]string) *Env {
	return &Env{
		original: ParseEnviron(os.Environ()),
		new:      env,
	}
}

// Export exports custom variables to the environment.
func (e *Env) Export() {
	for k, v := range e.new {
		os.Setenv(k, v)
	}
}

// Restore restores the otiginal environment.
func (e *Env) Restore() {
	for k := range e.new {
		os.Unsetenv(k)
	}

	for k, v := range e.original {
		os.Setenv(k, v)
	}
}
