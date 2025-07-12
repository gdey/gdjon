package refobj

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sync/atomic"

	"github.com/ohler55/ojg/jp"
)

type Object[T any] = *object[T]

type object[T any] struct {
	path     jp.Expr
	hydrated *atomic.Bool
	obj      *T

	// MarshalUnhydrated is set to true, when marshaling the datastructure it will return these values as string set to json paths, if that is what it was initially.
	MarshalUnhydrated bool
}

var jsonPathRegex = regexp.MustCompile(`^\${path:([^}]+)}$`)

func (obj *object[T]) UnmarshalJSON(data []byte) error {
	if obj.hydrated == nil {
		obj.hydrated = new(atomic.Bool)
	}

	if obj.obj == nil {
		obj.obj = new(T)
	}

	hydrated, err := unmarshal(data, obj.obj, func(m string, isStr bool) (err error) {
		idx := jsonPathRegex.FindSubmatchIndex([]byte(m))
		if len(idx) == 0 {
			return fmt.Errorf("bad path expression: %v", m)
		}
		jsonPath := "$." + m[idx[2]:idx[3]]
		obj.path, err = jp.ParseString(jsonPath)
		if err != nil {
			return fmt.Errorf("parse error: %w", err)
		}
		return nil

	})
	if err != nil {
		return err
	}
	obj.hydrated.Store(hydrated)
	return nil
}

func (obj *object[T]) MarshalJSON() ([]byte, error) {
	if obj == nil {
		return nil, nil
	}
	if obj.MarshalUnhydrated && obj.path != nil {
		// need to strip out `$.`
		path := obj.path.String()

		return []byte(`${path:` + path[2:] + `}`), nil
	}
	return json.Marshal(obj.obj)
}

func (obj *object[T]) Hydrate(data any) error {
	if obj == nil {
		return nil
	}
	if obj.hydrated == nil {
		obj.hydrated = new(atomic.Bool)
	} else if obj.hydrated.Load() {
		return nil
	}

	results := obj.path.Get(data)
	if len(results) == 0 {
		return fmt.Errorf("not found")
	}

	item := results[0]
	if a, ok := item.(Hydrator); ok {
		// Make sure to hydrate sub items
		a.Hydrate(data)
	}

	// if obj.obj is json.Rawmessage we need to deal with it differently.
	if _, ok := any(obj.obj).(*json.RawMessage); ok {
		data, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("failed to marshal for raw message: %w", err)
		}
		err = json.Unmarshal(data, obj.obj)
		if err != nil {
			return err
		}
		obj.hydrated.Store(true)
		return nil
	}

	// Generally expecting only one object
	// TODO(gdey): work with more then one result, where T is []U so, we convert []any to []U.
	val, ok := item.(T)
	if !ok {
		return fmt.Errorf("unexpected object type: expected %T got %T", obj.obj, results[0])
	}
	obj.obj = &val
	obj.hydrated.Store(true)
	return nil
}

func (obj object[T]) Object() (val T) {
	if obj.hydrated == nil || !obj.hydrated.Load() || obj.obj == nil {
		return
	}
	return *obj.obj
}
func (obj object[T]) Hydrated() bool {
	if obj.hydrated == nil {
		return false
	}
	return obj.hydrated.Load()
}
