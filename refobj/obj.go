package refobj

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sync/atomic"

	"github.com/davecgh/go-spew/spew"
	"github.com/ohler55/ojg/jp"
)

type Object[T any] = *object[T]

type object[T any] struct {
	path     jp.Expr
	hydrated *atomic.Bool
	obj      *T
}

var jsonPathRegex = regexp.MustCompile(`^\${path:([^}]+)}$`)

func (obj *object[T]) UnmarshalJSON(data []byte) error {
	if obj.hydrated == nil {
		obj.hydrated = new(atomic.Bool)
	}

	log.Printf("obj UnmarshalJSON data: %v", data)
	if obj.obj == nil {
		obj.obj = new(T)
	}

	hydrated, err := unmarshal(data, obj.obj, func(m string, isStr bool) (err error) {
		idx := jsonPathRegex.FindSubmatchIndex([]byte(m))
		if len(idx) == 0 {
			return fmt.Errorf("bad path expression: %v", m)
		}
		log.Printf("idx: %v: %s", idx, m[idx[2]:idx[3]])
		jsonPath := "$." + m[idx[2]:idx[3]]
		log.Printf("JsonPath: %s", jsonPath)
		obj.path, err = jp.ParseString(jsonPath)
		if err != nil {
			log.Printf("Got error parsing: %s : %v", jsonPath, err)
			return fmt.Errorf("parse error: %w", err)
		}
		log.Printf("Returning Nil for err")
		return nil

	})
	log.Printf("unmarshal returned an error: %v", err)
	if err != nil {
		return err
	}
	obj.hydrated.Store(hydrated)
	return nil
}

func (obj *object[T]) Hydrate(data any) error {
	if obj == nil {
		log.Printf("object is nil")
		return nil
	}
	if obj.hydrated == nil {
		obj.hydrated = new(atomic.Bool)
	} else if obj.hydrated.Load() {
		log.Printf("object hydrated is true")
		return nil
	}

	results := obj.path.Get(data)
	log.Printf("object get (%v): %v", obj.path, spew.Sdump(results))
	if len(results) == 0 {
		log.Printf("Returning ErrNotExist")
		return os.ErrNotExist
	}

	log.Printf("Attempting to set value")
	// Generally expecting only one object
	// TODO(gdey): work with more then one result, where T is []U so, we convert []any to []U.
	val, ok := results[0].(T)
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
