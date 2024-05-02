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
	Values(any)
}
