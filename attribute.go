package goe

import (
	"fmt"
)

type att struct {
	table string
	name  string
}

func (a *att) Equals(v any) boolean {
	fmt.Println(a.table + "." + a.name + " = " + "$1")
	return boolean{}
}

// func UUID(name string) Attribute {
// 	return &attribute{Name: name, Type: "UUID"}
// }
