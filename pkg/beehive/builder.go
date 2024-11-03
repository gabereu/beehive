package beehive

import (
	"reflect"
)

type BeeBuilder[T any] struct {
	beeType reflect.Type
	data    beeData
}

func NewBuilder[T any]() BeeBuilder[T] {
	return BeeBuilder[T]{
		beeType: reflect.TypeFor[T](),
		data:    beeData{},
	}
}

func (b BeeBuilder[T]) Name(name any) BeeBuilder[T] {
	b.data.name = name

	return b
}

func (b BeeBuilder[T]) Value(resource T) BeeBuilder[T] {
	b.data.created = &resource

	return b
}

func (b BeeBuilder[T]) Deps(deps ...beeId) BeeBuilder[T] {
	b.data.deps = deps

	return b
}

func (b BeeBuilder[T]) Func(creation any) BeeBuilder[T] {
	b.data.creation = creation
	return b
}

func (b BeeBuilder[T]) Register(hive *Hive) error {

	if b.data.creation != nil && b.data.deps == nil {
		typeF := reflect.TypeOf(b.data.creation)
		numIn := typeF.NumIn()
		deps := make(Deps, numIn)
		for i := 0; i < numIn; i++ {
			depType := typeF.In(i)
			deps[i] = beeId{type_: depType, name: nil}
		}
		b.data.deps = deps
	}

	return hive.register(b.beeType, b.data)
}

func Id[T any](name Name) beeId {
	beeType := reflect.TypeFor[T]()

	return beeId{
		type_: beeType,
		name:  name,
	}
}
