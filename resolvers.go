package submodule

import (
	"fmt"
	"reflect"
)

func isInEmbedded(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Type == inType {
			return true
		}
	}
	return false
}

func isSelf(t reflect.Type) bool {
	return t.AssignableTo(selfType)
}

func resolveEmbedded(as *store, t reflect.Type, v reflect.Value, dependencies []Retrievable) (reflect.Value, error) {
	var pt reflect.Type
	var pv reflect.Value

	store := getStore()
	if as != nil {
		store = as
	}

	if t.Kind() == reflect.Pointer {
		pv = reflect.Indirect(v)
		pt = t.Elem()
	} else {
		pv = reflect.Indirect(v)
		pt = t
	}

	for i := 0; i < pt.NumField(); i++ {
		f := pt.Field(i)
		if f.Type == inType {
			continue
		}

		if !f.IsExported() {
			return pv, fmt.Errorf("unable to resolve unexported field: %s, field is not exported", f.Name)
		}

		value, err := resolveType(store, f.Type, dependencies)
		if err != nil {
			return pv, err
		}

		pv.Field(i).Set(value)
	}

	if t.Kind() == reflect.Pointer {
		return pv.Addr(), nil
	}

	return pv, nil
}

func resolveType(store *store, t reflect.Type, dependencies []Retrievable) (v reflect.Value, e error) {
	if isInEmbedded(t) {
		var sv reflect.Value
		if t.Kind() == reflect.Pointer {
			sv = reflect.New(t.Elem())
		} else {
			sv = reflect.New(t)
		}

		v, e = resolveEmbedded(store, t, sv, dependencies)
		return
	}

	for _, d := range dependencies {
		if d.canResolve(t) {
			vv, err := d.retrieve(store)
			if err != nil {
				return v, err
			}

			v = reflect.ValueOf(vv)
			return
		}
	}
	return v, fmt.Errorf("unable to resolve dependency for type: %s", t.String())
}
