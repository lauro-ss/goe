package tests_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/lauro-ss/goe"
	"github.com/lauro-ss/postgres"
)

type Animal struct {
	Id        int
	Name      string
	IdHabitat *uuid.UUID `goe:"table:Habitat"`
	IdInfo    *[]byte    `goe:"table:Info"`
	Foods     []Food     `goe:"table:AnimalFood"`
}

type Food struct {
	Id      int
	Name    string
	Animals []Animal `goe:"table:AnimalFood"`
}

type Habitat struct {
	Id      uuid.UUID
	Name    string `goe:"type:varchar(50)"`
	Animals []Animal
}

type Info struct {
	Id   []byte
	Name string
}

type Database struct {
	Animal  *Animal
	Food    *Food
	Habitat *Habitat
	Info    *Info
	*goe.DB
}

func SetupPostgres() (*Database, error) {
	db := &Database{DB: &goe.DB{}}
	err := goe.Open(db, postgres.Open("user=postgres password=postgres host=localhost port=5432 database=postgres"), goe.Config{})
	if err != nil {
		return nil, err
	}
	err = db.Migrate(goe.MigrateFrom(db))
	if err != nil {
		return nil, err
	}
	return db, nil
}

func TestPostgresConection(t *testing.T) {
	_, err := SetupPostgres()
	if err != nil {
		t.Fatalf("Expetec Postgres Conection, got error %v", err)
	}
}
