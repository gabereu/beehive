package beehive

import (
	"errors"
	"fmt"
	"reflect"
)

type Name any
type Deps []beeId
type beesNode map[Name]*beeData

type beeId struct {
	type_ reflect.Type
	name  Name
}

type beeData struct {
	name     Name
	deps     Deps
	creation any //must be func
	created  any
}

func (id beeId) String() string {
	if id.name == nil {
		return fmt.Sprintf("'%s'", id.type_)
	} else {
		return fmt.Sprintf("'%s#%s'", id.type_, id.name)
	}
}

func (data beeData) validate() error {
	if data.created == nil {
		if data.creation == nil {
			return errors.New("bee must has an value or an creation strategy")
		}

		if _, ok := data.creation.(reflect.Type); ok {
			return nil
		}

		err := validateCreationFunc(data)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateCreationFunc(data beeData) error {

	typeF := reflect.TypeOf(data.creation)
	numIn := typeF.NumIn()

	if len(data.deps) != numIn {
		return fmt.Errorf("incorrect number of dependencies %d mus be %d", len(data.deps), numIn)
	}

	depsArgs := make([]reflect.Type, numIn)
	realArgs := make([]reflect.Type, numIn)
	for i := 0; i < numIn; i++ {
		realArgs[i] = typeF.In(i)
		depsArgs[i] = data.deps[i].type_
	}

	if !reflect.DeepEqual(depsArgs, realArgs) {
		id := beeId{
			type_: typeF.In(0),
			name:  data.name,
		}
		return fmt.Errorf("bee %s creation func with invalid parameters %v for real parameters %v", id, depsArgs, realArgs)
	}

	return nil
}
