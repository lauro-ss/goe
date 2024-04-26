package goe

// type Database interface {
// 	Select(...Attribute) Rows
// }

type Table interface {
	Select(...any) Rows
	//Join(string) Join
	// Update() int
	// Delete() int
	// Create() int
}

type Where interface {
	Where(...*booleanResult) Rows
}

type Join interface {
	Join(any) Rows
}

type Rows interface {
	Where
	//Join
	Result
}

type Result interface {
	Result(target any)
}
