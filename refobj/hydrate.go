package refobj

import (
	"encoding/json"
	"log"
	"reflect"

	"github.com/davecgh/go-spew/spew"
)

type Hydrator interface {
	Hydrate(data any) error
}

func hydrate(top, val any) (err error) {
	log.Printf("called Hydrate %T", val)

	// let's first see if val is Hydratable?
	if h, ok := val.(Hydrator); ok {
		log.Printf("calling Hydrate interface")
		h.Hydrate(top)
	}

	// now I need to iterate through val to figure out if things need to be Hydrated
	tv := reflect.TypeOf(val)
	vv := reflect.ValueOf(val)
restart:
	log.Printf("Looking at kind: %v ", tv.Kind())
	switch tv.Kind() {
	case reflect.Pointer:
		vv = vv.Elem()
		tv = vv.Type()
		goto restart

	case reflect.Array:
		log.Printf("Hydrate array")
		for _, v := range vv.Seq2() {
			err = hydrate(top, v.Interface())
			if err != nil {
				return err
			}
		}

	case reflect.Slice:
		log.Printf("Hydrate slice")
		for _, v := range vv.Seq2() {
			err = hydrate(top, v.Interface())
			if err != nil {
				return err
			}
		}

	case reflect.Map:
		log.Printf("Hydrate map")
		iter := vv.MapRange()
		for iter.Next() {
			err = hydrate(top, iter.Value().Interface())
			if err != nil {
				return err
			}
		}

	case reflect.Struct:
		log.Printf("Hydrate struct")
		for i := range vv.NumField() {
			vf := vv.Field(i)
			name := tv.Field(i).Name
			vAny := vf.Interface()
			log.Printf("Struct: %s", spew.Sdump(vAny))
			log.Printf("Attempting to hydrate %v (%T) field (%v: %v)", name, vAny, vf.CanSet(), vf.CanAddr())
			if h, ok := vAny.(Hydrator); ok {
				log.Printf("Hydrating directly: %v", name)
				h.Hydrate(top)
				continue
			}
			if vf.CanSet() {
				log.Printf("Hydrating struct field: %v", name)
				err = hydrate(top, vf.Interface())
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func UnmarshalJSON[T any](data []byte, val *T) error {
	// first unmarshal it all into val
	err := json.Unmarshal(data, val)
	if err != nil {
		log.Printf("WE got an error: %v -- %s : %v", err, data, val)
		return err
	}

	//return nil
	return hydrate(val, val)
}

func Hydrate(top, val any) error { return hydrate(top, val) }
