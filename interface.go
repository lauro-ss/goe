package goe

type StateSelect interface {
	Where
	Result
}

type Where interface {
	Where(...*booleanResult) StateSelect
}

type Result interface {
	Result(target any)
}

type StateInsert interface {
	Value(any)
}

type StateBetwent interface {
	Values(any, any)
}
