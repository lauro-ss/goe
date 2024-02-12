package goe

type Database interface {
	Select(...string) From
	// Update() int
	// Delete() int
	// Create() int
}

type From interface {
	From(string) Table
}

type Table interface {
	Join(string) Join
	// Update() int
	// Delete() int
	// Create() int
}

type Join interface {
	Join(string) Join
	// Update() int
	// Delete() int
	// Create() int
}
