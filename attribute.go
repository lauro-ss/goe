package goe

import (
	"fmt"
)

type Pk interface {
	attribute
}

type Att interface {
	attribute
}

type pk struct {
	table string
	name  string
	Fk    map[string]*pk
}

func (a *pk) Equals(v any) boolean {
	fmt.Println(a.table + "." + a.name + " = " + "$1")
	return boolean{}
}

type att struct {
	name string
	pk   *pk
}

func (a *att) Equals(v any) boolean {
	// fmt.Println(a.table + "." + a.name + " = " + "$1")
	return boolean{}
}

// func UUID(name string) Attribute {
// 	return &attribute{Name: name, Type: "UUID"}
// }
