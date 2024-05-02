package submodule

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var intValue = Make[int](func() int {
	return 100
})

var stringValue = Make[string](func() string {
	return "hello"
})

type Embedded struct {
	In
	Int    int
	String string
}

func TestResolver(t *testing.T) {
	t.Run("can resolve type", func(t *testing.T) {
		store := CreateStore()
		r, e := resolveType(store, reflect.TypeOf(0), []Retrievable{intValue})
		assert.Nil(t, e)
		assert.Equal(t, r.Interface(), 100)

		r, e = resolveType(
			store,
			reflect.TypeOf(Embedded{}),
			[]Retrievable{intValue, stringValue},
		)

		v, ok := r.Interface().(Embedded)
		assert.True(t, ok)
		assert.Nil(t, e)
		assert.Equal(t, 100, v.Int)
		assert.Equal(t, "hello", v.String)
	})

	t.Run("can resolve embedded type", func(t *testing.T) {
		store := CreateStore()

		var v Embedded

		_, e := resolveEmbedded(
			store,
			reflect.TypeOf(Embedded{}),
			reflect.ValueOf(&v),
			[]Retrievable{intValue, stringValue},
		)

		assert.Nil(t, e)

		assert.Equal(t, v.Int, 100)
		assert.Equal(t, v.String, "hello")
	})

	t.Run("value can be replaced", func(t *testing.T) {
		store := CreateStore()
		store.InitValue(intValue, 200)

		var v Embedded

		_, e := resolveEmbedded(
			store,
			reflect.TypeOf(Embedded{}),
			reflect.ValueOf(&v),
			[]Retrievable{intValue, stringValue},
		)

		assert.Nil(t, e)

		assert.Equal(t, v.Int, 200)
		assert.Equal(t, v.String, "hello")
	})

}
