package goe

import "fmt"

type simpleOperator struct {
	operator string
}

func (so simpleOperator) operation() string {
	return fmt.Sprintf(" %v ", so.operator)
}

type complexOperator struct {
	argument  string
	operator  string
	valueFlag string
	value     any
	pk        *pk
}

func createComplexOperator(argument string, operator string, value any, pk *pk) complexOperator {
	return complexOperator{argument: argument, operator: operator, value: value, pk: pk}
}

func (co complexOperator) operation() string {
	return fmt.Sprintf("%v %v %v", co.argument, co.operator, co.valueFlag)
}

func (co *complexOperator) setValueFlag(f any) {
	co.valueFlag = f.(string)
}
