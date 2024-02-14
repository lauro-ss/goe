package goe

import (
	"fmt"
	"reflect"
)

type Database interface {
	Select(...Attribute) Where
	// Update() int
	// Delete() int
	// Create() int
}

type Attribute struct {
	Table string
	Name  string
	Type  string
}

func (a *Attribute) Equals(v any) where {
	fmt.Println(a.Table + "." + a.Name + " = " + reflect.ValueOf(v).String())
	return where{}
}

type where struct{}

type Where interface {
	Where(where)
}

type From interface {
	From(string) Table
}

type Table interface {
	Join(string) Join
	// Update() int
	// Delete() int
	// Create() int
}

type Join interface {
	Join(string) Join
	// Update() int
	// Delete() int
	// Create() int
}
