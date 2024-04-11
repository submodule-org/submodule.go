package submodule

import (
	"fmt"
	"reflect"
)

type in struct{}
type In interface {
	IsIn() bool
}

func (i in) IsIn() bool { return true }

var inType = reflect.TypeOf(new(In)).Elem()

func isInEmbed(t reflect.Type) bool {
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

func resolveEmbedded(st any, dependencies []gettable) (v any, e error) {
	var t reflect.Type
	var sv reflect.Value

	if reflect.TypeOf(st).Kind() == reflect.Pointer {
		t = reflect.TypeOf(st).Elem()
		sv = reflect.ValueOf(st).Elem()
	} else {
		t = reflect.TypeOf(st)
		sv = reflect.ValueOf(st)
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Type == inType {
			continue
		}

		if !f.IsExported() {
			return v, fmt.Errorf("unable to resolve unexported field: %s", f.Type.String())
		}

		value, err := resolveType(f.Type, dependencies, &f)
		if err != nil {
			return v, err
		}

		sv.Field(i).Set(value)
	}
	return sv.Interface(), nil
}

func resolveType(t reflect.Type, dependencies []gettable, s *reflect.StructField) (v reflect.Value, e error) {

	groupLooking := false

	if s != nil {
		groupLooking = s.Tag.Get("group") == "true"
	}

	if groupLooking && t.Kind() != reflect.Slice {
		return v, fmt.Errorf("field: %s type must be a slice to look up for group", t.Name())
	}

	if !groupLooking {
		for _, d := range dependencies {
			if d.CanResolve(t.Name()) {
				vv, err := d.Get()
				if err != nil {
					return
				}

				v = reflect.ValueOf(vv)
				return
			}
		}
		return v, fmt.Errorf("unable to resolve dependency for type: %s", t.String())
	} else {
		xt := reflect.MakeSlice(t, 0, 0)

		typeToLookup := xt.Elem().Type().Name()
		for _, d := range dependencies {
			if d.CanResolve(typeToLookup) {
				vv, err := d.Get()
				if err != nil {
					return
				}

				v = reflect.ValueOf(vv)
				xt = reflect.AppendSlice(xt, v)
			}
		}

		return xt, nil
	}
}

func resolveTypes(types []reflect.Type, dependencies []gettable) ([]reflect.Value, error) {

	args := make([]reflect.Value, len(types))
	for i := 0; i < len(types); i++ {

		if isInEmbed(types[i]) {
			v, e := resolveEmbedded(reflect.New(types[i]).Interface(), dependencies)
			if e != nil {
				return nil, fmt.Errorf("unable to resolve embedded type: %s, %w", types[i].String(), e)
			}
			args[i] = reflect.ValueOf(v)
			continue
		}

		v, error := resolveType(types[i], dependencies, nil)
		if error != nil {
			return nil, error
		}

		args[i] = v
	}

	return args, nil
}

func Provide[T any](fn func() T, alters ...alterable) Submodule[T] {
	return construct[T](fn, alters...)
}

func ProvideWithError[T any](fn func() (T, error), alters ...alterable) Submodule[T] {
	return construct[T](fn, alters...)
}

func Make[T any](fn any, dependencies ...alterable) Submodule[T] {
	return construct[T](fn, dependencies...)
}

func Craft[T any](t T, dependencies ...alterable) Submodule[T] {
	tt := reflect.TypeOf(t)

	if tt.Kind() != reflect.Struct && tt.Kind() != reflect.Pointer && tt.Kind() != reflect.Func {
		panic(fmt.Sprintf("only struct or struct pointer or func: %v", tt.String()))
	}

	gettables, _ := extract(dependencies)

	return construct[T](func() T {
		_, e := resolveEmbedded(t, gettables)
		if e != nil {
			e = fmt.Errorf("unable to resolve embedded type: %s, %w", tt.String(), e)
			panic(e)
		}

		return t
	}, dependencies...)
}

func Construct[T any, D In](fn func(d D) T, dependencies ...alterable) Submodule[T] {
	return construct[T](fn, dependencies...)
}

func Override[T any](s Submodule[T], dependencies ...alterable) {
	sm := s.(*submodule[T])
	var nds []gettable
	gettables, _ := extract(dependencies)

	nds = append(nds, gettables...)
	nds = append(nds, sm.valuer.dependencies...)

	sm.valuer.dependencies = nds
}
