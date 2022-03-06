package main

import (
	"context"
	"fmt"
	"reflect"
)

func main() {
	ctx := context.Background()
	r := reflect.ValueOf(ctx)
	fmt.Println("xxxxxx", r.Type())
}
