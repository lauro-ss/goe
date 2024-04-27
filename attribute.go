package goe

import "fmt"

type manyToMany struct {
	table string
	ids   map[string]string
}

type manyToOne struct {
	id      string
	hasMany bool
}

type pk struct {
	table    string
	name     string
	skipFlag bool
	fks      map[string]any
}

type att struct {
	name string
	pk   *pk
}

const (
	EQUALS = 1
)

type booleanResult struct {
	arg   string
	pk    *pk
	value string
	tip   int8
}

func createBooleanResult(arg string, pk *pk, value any, tip int8) *booleanResult {
	return &booleanResult{arg: arg, pk: pk, value: fmt.Sprintf("'%v'", value), tip: tip}
}
