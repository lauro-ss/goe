package goe

type Select interface {
	WhereSelect
	Join
	Result
}

type WhereSelect interface {
	Where(...operator) SelectWhere
}

type SelectWhere interface {
	Result
}

type Join interface {
	Join(...any) Select
}

type Result interface {
	Result(any)
}

type Insert interface {
	Value
}

type InsertBetwent interface {
	Values
}

type Update interface {
	WhereUpdate
	Value
}

type WhereUpdate interface {
	Where(...operator) UpdateWhere
}

type UpdateWhere interface {
	Value
}

type Value interface {
	Value(any)
}

type Values interface {
	Values(any, any)
}

type Delete interface {
	Where(...operator)
}

type operator interface {
	operation() string
}
