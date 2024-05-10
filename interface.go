package goe

type Select interface {
	WhereSelect
	Result
}

type WhereSelect interface {
	Where(...*booleanResult) SelectWhere
}

type SelectWhere interface {
	Result
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
	Where(...*booleanResult) UpdateWhere
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
	Where(...*booleanResult)
}
