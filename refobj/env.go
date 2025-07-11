package refobj

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

type Env[T ~string] = *env[T]

// env is restricted to string types cause don't want to figure
// out how to deal with other types when an env is only string
type env[T ~string] struct {
	// MarshalUnhydrated is set to true, when marshaling the datastructure it will return these values as string set to json paths, if that is what it was initially.
	MarshalUnhydrated bool
	// envVar is the env var name
	envVar string
	obj    *T
}

var envPathRegex = regexp.MustCompile(`^\${env:([^}]+)}$`)

func (e *env[T]) UnmarshalJSON(data []byte) (err error) {
	if e.obj == nil {
		e.obj = new(T)
	}
	_, err = unmarshal(data, e.obj, func(m string, isStr bool) (err error) {
		idx := envPathRegex.FindSubmatchIndex([]byte(m))
		if len(idx) == 0 {
			return fmt.Errorf("bad env expression: %v", m)
		}
		e.envVar = m[idx[2]:idx[3]]
		*e.obj = T(os.Getenv(e.envVar))
		return nil
	})
	return err
}

func (e *env[T]) MarshalJSON() ([]byte, error) {
	if e == nil {
		return nil, nil
	}
	if e.MarshalUnhydrated && e.envVar != "" {
		return []byte(`${env:` + e.envVar + `}`), nil
	}
	return json.Marshal(e.obj)
}

func (*env[T]) Hydrated(any) error { return nil }
func (e *env[T]) Object() T {
	if e == nil || e.obj == nil {
		return T("")
	}
	return *e.obj
}
