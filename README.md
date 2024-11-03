<img src="assets/beehive.png" alt="Beehive logo" width="150">

# Beehive
Beehive is a golang **dependency injection** container.
Designed to be simple, easy to use and read.

## Usage

add to you project

```bash
go get github.com/gabereu/beehive
```

register your bees (resources) and get then when you need

```go
import (
	"fmt"

	"github.com/gabereu/beehive"
)

func main() {
	hive := beehive.NewHive()
	hive.Register("Some string value")

	var byPointer string
	err := hive.Get(&byPointer)

	byCall, err2 := beehive.Get[string](hive)

	fmt.Printf("byPointer %s \nerr %v \nbyCall %s \nerr2 %v", byPointer, err, byCall, err2)
}
```

## Features

- register already created resources
    ```go
	hive.Register("Some string value")
	hive.RegisterWithName("some name", "another string value")
    ```
- register functions that create the resource and receive the required dependencies to it
  ```go
	hive.RegisterFunc(func(value int) string {
		return fmt.Sprintf("received %d", value)
	})
    hive.RegisterFuncWithName(
        "received.other", 
        func(value int) string {
		    return fmt.Sprintf("received with name %d", value)
	    })
    
    // if function deps has names use Builder to define dependencies ids
    beehive.NewBuilder[string]().
        Name("name of func resource").
        Deps(beehive.Id[int]("config.value")).
        Func(func(configValue int) string {
            return fmt.Sprintf("from builder %d", configValue)
        }).
        Register(hive)
    ```
- register structs that hive will create and inject the dependencies in the fields (so the passed instance will not be used)
    ```go

    type Service struct {
		confing int `bee:""`
		config2 string
		config3 string `bee:"config.3"`
	}

    // Register struct value
	err := hive.RegisterStruct(Service{})
	service, err := beehive.Get[Service](hive)

    // Or register struct pointer
	err := hive.RegisterStruct(&Service{})
	service, err := beehive.Get[*Service](hive) // when getting bee: Service != *Service

    // Use with generics to avoid thinking the variable will be used
	err := beehive.RegisterStruct[*Service](hive)
	service, err := beehive.Get[*Service](hive)
    ```
