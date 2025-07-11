package refobj

import (
	"encoding/json"
	"reflect"
)

type Hydrator interface {
	Hydrate(data any) error
}

func hydrate(top, val any) (err error) {

	// let's first see if val is Hydratable?
	if h, ok := val.(Hydrator); ok {
		h.Hydrate(top)
	}

	// now I need to iterate through val to figure out if things need to be Hydrated
	tv := reflect.TypeOf(val)
	vv := reflect.ValueOf(val)
restart:
	switch tv.Kind() {
	case reflect.Pointer:
		vv = vv.Elem()
		tv = vv.Type()
		goto restart

	case reflect.Array:
		for _, v := range vv.Seq2() {
			err = hydrate(top, v.Interface())
			if err != nil {
				return err
			}
		}

	case reflect.Slice:
		for _, v := range vv.Seq2() {
			err = hydrate(top, v.Interface())
			if err != nil {
				return err
			}
		}

	case reflect.Map:
		iter := vv.MapRange()
		for iter.Next() {
			err = hydrate(top, iter.Value().Interface())
			if err != nil {
				return err
			}
		}

	case reflect.Struct:
		for i := range vv.NumField() {
			vf := vv.Field(i)
			vAny := vf.Interface()
			if h, ok := vAny.(Hydrator); ok {
				h.Hydrate(top)
				continue
			}
			if vf.CanSet() {
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
	if err := json.Unmarshal(data, val); err != nil {
		return err
	}

	return hydrate(val, val)
}

func Hydrate(top, val any) error { return hydrate(top, val) }
