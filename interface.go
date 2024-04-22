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

type attribute interface {
	Equals(v any) boolean
}

type Where interface {
	Where(boolean) Rows
}

type boolean struct{}

type Join interface {
	Join(string) Join
}

type From interface {
	From(any) Rows
}

type Rows interface {
	//Where
	//Join
	Result
}

type Result interface {
	Result(target any)
}
