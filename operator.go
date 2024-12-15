package goe

import (
	"fmt"
	"reflect"
)

type simpleOperator struct {
	operator string
}

func (so simpleOperator) operation() string {
	return fmt.Sprintf(" %v ", so.operator)
}

type fieldOperator struct {
	argument string
	operator string
	field    string
}

func (fo fieldOperator) operation() string {
	return fmt.Sprintf("%v %v %v", fo.argument, fo.operator, fo.field)
}

type complexOperator struct {
	argument  string
	operator  string
	valueFlag string
	value     any
	pk        *pk
}

func createComplexOperator(argument string, operator string, value any, pk *pk, addrMap map[uintptr]field) operator {
	valueOf := reflect.ValueOf(value)
	if valueOf.Kind() == reflect.Pointer {
		ptr := uintptr(valueOf.Elem().Addr().UnsafePointer())
		if f := addrMap[ptr]; f != nil {
			return fieldOperator{argument: argument, operator: operator, field: f.getSelect()}
		}
	}
	return complexOperator{argument: argument, operator: operator, value: value, pk: pk}
}

func (co complexOperator) operation() string {
	return fmt.Sprintf("%v %v %v", co.argument, co.operator, co.valueFlag)
}

func (co *complexOperator) setValueFlag(f string) {
	co.valueFlag = f
}

func (co *complexOperator) setNot() {
	co.operator = "NOT " + co.operator
}
