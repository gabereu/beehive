package beehive

import (
	"fmt"
	"reflect"
	"strings"
)

type Hive struct {
	deps map[reflect.Type]beesNode
}

func NewHive() *Hive {
	deps := make(map[reflect.Type]beesNode)
	return &Hive{deps}
}

func (h *Hive) register(t reflect.Type, bee beeData) error {

	err := bee.validate()
	if err != nil {
		return err
	}

	bees, ok := h.deps[t]
	if !ok {
		bees = map[Name]*beeData{}
		h.deps[t] = bees
	} else {
		_, ok = bees[bee.name]
		if ok {
			id := beeId{type_: t, name: bee.name}
			return fmt.Errorf("bee %s already registered", id)
		}
	}

	bees[bee.name] = &bee
	return nil
}

func (h *Hive) Register(bee any) error {
	beeType := reflect.TypeOf(bee)

	data := beeData{
		name:    nil,
		created: &bee,
	}

	return h.register(beeType, data)
}

func (h *Hive) RegisterWithName(name Name, bee any) error {
	beeType := reflect.TypeOf(bee)

	data := beeData{
		name:    name,
		created: &bee,
	}

	return h.register(beeType, data)
}

func (h *Hive) RegisterFunc(creation any) error {
	return h.RegisterFuncWithName(nil, creation)
}

func (h *Hive) RegisterFuncWithName(name Name, creation any) error {

	creationType := reflect.TypeOf(creation)

	if creationType.Kind() != reflect.Func {
		return fmt.Errorf("you must provide an function as creation argument")
	}

	if creationType.NumOut() != 1 {
		isSecondReturnError := creationType.NumOut() == 2 && creationType.Out(1).AssignableTo(reflect.TypeFor[error]())
		if !isSecondReturnError {
			return fmt.Errorf("you must provide an function that return only one value or return (any, error)")
		}
	}

	beeType := creationType.Out(0)

	deps := make(Deps, creationType.NumIn())
	for i := 0; i < creationType.NumIn(); i++ {
		inType := creationType.In(i)
		deps[i] = beeId{
			type_: inType,
			name:  nil,
		}
	}

	bee := beeData{
		name:     name,
		deps:     deps,
		creation: creation,
	}

	return h.register(beeType, bee)
}

func Get[T any](hive *Hive) (T, error) {
	var value T
	err := hive.Get(&value)
	return value, err
}

func GetByName[T any](hive *Hive, name Name) (T, error) {
	var value T
	err := hive.GetByName(name, &value)
	return value, err
}

func (h *Hive) Get(ptrToBee any) error {
	return h.GetByName(nil, ptrToBee)
}

func (h *Hive) GetByName(name Name, ptrToValue any) error {

	id := beeId{
		name:  name,
		type_: reflect.TypeOf(ptrToValue).Elem(),
	}

	visited := []beeId{id}
	content, err := h.get(id, visited)
	if err != nil {
		return err
	}

	resource := h.deps[id.type_][id.name]
	resource.created = &content
	setValues(resource.created, ptrToValue)

	return nil
}

func (h *Hive) get(id beeId, visited []beeId) (any, error) {

	var empty any

	bees, ok := h.deps[id.type_]
	if !ok {
		return empty, fmt.Errorf("bee %s not found", id)
	}

	bee, ok := bees[id.name]
	if !ok {
		return empty, fmt.Errorf("bee %s not found", id)
	}

	if bee.created != nil {
		created := reflect.ValueOf(bee.created).Elem().Interface()
		return created, nil
	}

	deps := make([]reflect.Value, len(bee.deps))
	for i, depId := range bee.deps {
		for _, v := range visited {
			if v == depId {
				return nil, fmt.Errorf("dependecy cycle found %v <-> %v | route: %s", depId, id, h.printCycle(visited))
			}
		}

		visited = append(visited, depId)

		dep, err := h.get(depId, visited)
		if err != nil {
			return nil, err
		}
		deps[i] = reflect.ValueOf(dep)

		visited = visited[:len(visited)-1]
	}

	r, err := safeCall(id, bee.creation, deps)

	if err != nil {
		return nil, err
	}

	if len(r) > 1 && !r[1].IsNil() {
		return nil, r[1].Interface().(error)
	}

	result := r[0].Interface()

	bee.created = &result

	return result, nil
}

func (h *Hive) printCycle(cycle []beeId) string {

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%v", cycle[0]))
	for i := 1; i < len(cycle); i++ {
		sb.WriteString(fmt.Sprintf(" > %v", cycle[i]))
	}
	sb.WriteString(fmt.Sprintf("> %v", cycle[0]))

	return sb.String()
}

func setValues(source, target any) {
	targetValue := reflect.ValueOf(target).Elem()
	sourceValue := reflect.ValueOf(source).Elem().Elem()

	targetValue.Set(sourceValue)
}

func safeCall(id beeId, fn any, args []reflect.Value) (result []reflect.Value, err error) {

	funcValue := reflect.ValueOf(fn)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error in creation func for bee %s: %v", id, r)
		}
	}()

	// Call the function
	return funcValue.Call(args), nil
}
