package goe

import (
	"fmt"
)

type Pk struct {
	table string
	name  string
	Fk    map[string]*Pk
}

func (a *Pk) Equals(v any) boolean {
	fmt.Println(a.table + "." + a.name + " = " + "$1")
	return boolean{}
}

type Att struct {
	name string
	pk   *Pk
}

func (a Att) Equals(v any) boolean {
	// fmt.Println(a.table + "." + a.name + " = " + "$1")
	return boolean{}
}

// func UUID(name string) Attribute {
// 	return &attribute{Name: name, Type: "UUID"}
// }
