package beehive_test

import (
	"errors"
	"fmt"
	"testgo/pkg/beehive"
	"testing"

	"github.com/stretchr/testify/assert"
)

type S2 struct {
	val any
}

type S struct {
	val any
}

func (s *S) Get() any {
	return s.val
}

type I interface {
	Get() any
}

func TestAlreadyCreatedBees(t *testing.T) {
	hive := beehive.NewHive()

	err1 := hive.Register("some string")
	err2 := hive.RegisterWithName("any name", "another string")

	err3 := hive.Register("some string")
	err4 := hive.RegisterWithName("any name", "another string")

	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Equal(t, "bee 'string' already registered", err3.Error())
	assert.Equal(t, "bee 'string#any name' already registered", err4.Error())

	var v1 string
	err5 := hive.Get(&v1)
	assert.Nil(t, err5)
	assert.Equal(t, "some string", v1)

	var v2 string
	err6 := hive.GetByName("any name", &v2)
	assert.Nil(t, err6)
	assert.Equal(t, "another string", v2)

	v3, err7 := beehive.Get[string](hive)
	assert.Nil(t, err7)
	assert.Equal(t, "some string", v3)

	v4, err5 := beehive.GetByName[string](hive, "any name")
	assert.Nil(t, err5)
	assert.Equal(t, "another string", v4)
}

func TestGetError(t *testing.T) {
	hive := beehive.NewHive()
	hive.Register(1)

	_, err1 := beehive.Get[string](hive)
	_, err2 := beehive.GetByName[int](hive, "name")

	assert.Equal(t, "bee 'string' not found", err1.Error())
	assert.Equal(t, "bee 'int#name' not found", err2.Error())
}

func TestFunctionCreatedBees(t *testing.T) {
	t.Run("Unnamed_Bees", func(t *testing.T) {
		hive := beehive.NewHive()

		hive.Register(123)
		hive.RegisterFunc(func(i int) string {
			return fmt.Sprintf("func(%d)", i)
		})

		var v1 string
		err1 := hive.Get(&v1)
		assert.Nil(t, err1)
		assert.Equal(t, "func(123)", v1)
	})

	t.Run("Named_Bees", func(t *testing.T) {
		hive := beehive.NewHive()

		hive.Register(123)
		hive.RegisterFunc(func(i int) (string, error) {
			return fmt.Sprintf("func(%d)", i), nil
		})
		hive.RegisterFuncWithName("alternative", func(i int) string {
			return fmt.Sprintf("alternative func(%d)", i)
		})

		var v1 string
		err1 := hive.Get(&v1)
		assert.Nil(t, err1)
		assert.Equal(t, "func(123)", v1)

		var v2 string
		err2 := hive.GetByName("alternative", &v2)
		assert.Nil(t, err2)
		assert.Equal(t, "alternative func(123)", v2)
	})

	t.Run("func_with_error_in_return", func(t *testing.T) {
		hive := beehive.NewHive()

		hive.RegisterFuncWithName(
			"error",
			func() (string, error) {
				return "", errors.New("error creating string")
			})

		hive.RegisterFuncWithName(
			"no error",
			func() (string, error) {
				return "value of string", nil
			})

		_, err1 := beehive.GetByName[string](hive, "error")
		assert.Equal(t, "error creating string", err1.Error())
		value, err2 := beehive.GetByName[string](hive, "no error")
		assert.Nil(t, err2)
		assert.Equal(t, "value of string", value)

	})

	t.Run("func_with_error_in_register", func(t *testing.T) {
		hive := beehive.NewHive()

		err1 := hive.RegisterFunc("not function")

		err2 := hive.RegisterFuncWithName(
			"func with 2 args but 2nd is not error",
			func() (string, string) {
				return "", ""
			})

		err3 := hive.RegisterFuncWithName(
			"func with multiple args",
			func() (string, string, string) {
				return "", "", ""
			})

		assert.Equal(t, "you must provide an function as creation argument", err1.Error())
		assert.Equal(t, "you must provide an function that return only one value or return (any, error)", err2.Error())
		assert.Equal(t, "you must provide an function that return only one value or return (any, error)", err3.Error())
	})
}

func TestBuilder(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		hive := beehive.NewHive()

		hive.RegisterWithName(1, 10)
		hive.RegisterWithName(2, 20)
		hive.Register(30)

		err := beehive.NewBuilder[string]().
			Name("first").
			Deps(beehive.Id[int](1)).
			Func(func(i int) string {
				return fmt.Sprintf("from builder %d", i)
			}).
			Register(hive)

		err2 := beehive.NewBuilder[string]().
			Name("second").
			Deps(beehive.Id[int](2)).
			Func(func(i int) string {
				return fmt.Sprintf("from builder %d", i)
			}).
			Register(hive)

		err3 := beehive.NewBuilder[string]().
			Name("third").
			Func(func(i int) string {
				return fmt.Sprintf("from builder %d", i)
			}).
			Register(hive)

		err4 := beehive.NewBuilder[string]().
			Name("fourth").
			Value("from builder value 40").
			Register(hive)

		first, err5 := beehive.GetByName[string](hive, "first")
		second, err6 := beehive.GetByName[string](hive, "second")
		third, err7 := beehive.GetByName[string](hive, "third")
		fourth, err8 := beehive.GetByName[string](hive, "fourth")

		assert.Nil(t, err)
		assert.Nil(t, err2)
		assert.Nil(t, err3)
		assert.Nil(t, err4)
		assert.Nil(t, err5)
		assert.Nil(t, err6)
		assert.Nil(t, err7)
		assert.Nil(t, err8)

		assert.Equal(t, "from builder 10", first)
		assert.Equal(t, "from builder 20", second)
		assert.Equal(t, "from builder 30", third)
		assert.Equal(t, "from builder value 40", fourth)
	})

	t.Run("Error", func(t *testing.T) {
		hive := beehive.NewHive()
		err := beehive.NewBuilder[string]().
			Register(hive)

		err2 := beehive.NewBuilder[string]().
			Deps(beehive.Id[int](nil)).
			Func(func(int, string) bool {
				return true
			}).
			Register(hive)

		assert.Equal(t, "bee must has an value or an creation strategy", err.Error())
		assert.Equal(t, "incorrect number of dependencies 1 mus be 2", err2.Error())
	})
}

func TestDependencyCycle(t *testing.T) {
	hive := beehive.NewHive()

	beehive.NewBuilder[string]().
		Name("first").
		Deps(beehive.Id[string]("third")).
		Func(func(s string) string {
			return fmt.Sprintf("first > %s", s)
		}).
		Register(hive)

	beehive.NewBuilder[string]().
		Name("second").
		Deps(beehive.Id[string]("first")).
		Func(func(s string) string {
			return fmt.Sprintf("second > %s", s)
		}).
		Register(hive)

	beehive.NewBuilder[string]().
		Name("third").
		Deps(beehive.Id[string]("second")).
		Func(func(s string) string {
			return fmt.Sprintf("third > %s", s)
		}).
		Register(hive)

	_, err := beehive.GetByName[string](hive, "third")
	assert.Error(t, err)
	assert.Equal(t, "dependecy cycle found 'string#third' <-> 'string#first' | route: 'string#third' > 'string#second' > 'string#first'> 'string#third'", err.Error())
}

func TestFunctionCallErrors(t *testing.T) {
	t.Run("Incompatible_deps", func(t *testing.T) {
		hive := beehive.NewHive()

		hive.Register(1)
		hive.Register(float32(12.34))

		err := beehive.NewBuilder[string]().
			Deps(beehive.Id[string](nil), beehive.Id[float32](nil)).
			Func(func(int, float32) bool {
				return true
			}).
			Register(hive)

		assert.Equal(t, "bee 'int' creation func with invalid parameters [string float32] for real parameters [int float32]", err.Error())
	})

	t.Run("Panic_in_creation_func", func(t *testing.T) {
		hive := beehive.NewHive()

		hive.RegisterFunc(func() string {
			panic("some thing happen")
		})

		_, err := beehive.Get[string](hive)

		assert.Equal(t, "error in creation func for bee 'string': some thing happen", err.Error())
	})
}
