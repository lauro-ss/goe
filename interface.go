package goe

type State interface {
	Where
	Result
}

type Where interface {
	Where(...*booleanResult) State
}

type Result interface {
	Result(...any)
}
