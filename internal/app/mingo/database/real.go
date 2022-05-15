package database

import (
	"database/sql"

	"github.com/craigjperry2/mingo/internal/app/mingo"
	_ "github.com/mattn/go-sqlite3"
)

// Fake DB access layer for this iteration
type Db struct {
	*sql.DB
}

func NewRealDatabase() *Db {
	db, err := sql.Open("sqlite3", "/Users/craig/Code/github.com/craigjperry2/mingo/my.db")
	if err != nil {
		panic("Couldn't open my.db")
	}
	return &Db{db}
}

func (db *Db) Close()  {
	db.DB.Close()
}

func (db *Db) GetAll(offset int, limit int) ([]mingo.Person, error) {
	// Because of my HTMX UI impl, i can't get away with simple offset/limit pagination, so using cursor style instead
	var result []mingo.Person
	if rows, err := db.DB.Query(`select * from Person where Id > (?) limit (?);`, offset, limit); err != nil {
		panic(err)
	} else {
		for rows.Next() {
			var p mingo.Person
			err := rows.Scan(&p.Id, &p.Name, &p.Location)
			if err != nil {
				panic(err)
			}
			result = append(result, p)
		}
	}
	return result, nil
}

func (db *Db) Get(id int) (mingo.Person, error) {
	var result mingo.Person
	if err := db.DB.QueryRow(`select * from Person where Id = (?);`, id).Scan(&result.Id, &result.Name, &result.Location); err != nil {
		panic(err)
	}
	return result, nil
}

func (db *Db) Update(id int, name string, location string) (mingo.Person, error) {
	p := mingo.Person{id, name, location}
	if _, err := db.Exec(`UPDATE Person SET name = (?), location = (?) WHERE id = (?);`, name, location, id); err != nil {
		panic(err)
	}
	return p, nil
}

func (db *Db) Insert(name string, location string) (mingo.Person, error) {
	var p mingo.Person
	if result, err := db.Exec(`INSERT INTO Person (name, location) VALUES ((?), (?)) RETURNING id;`, name, location); err != nil {
		panic(err)
	} else {
		id, err := result.LastInsertId()
		if err != nil {
			panic(err)
		}
		p = mingo.Person{int(id), name, location}
	}
	return p, nil
}

func (db *Db) Delete(id int) (mingo.Person, error) {
	var old mingo.Person
	if _, err := db.Exec(`DELETE FROM Person WHERE id = (?);`, id); err != nil {
		panic(err)
	}
	return old, nil
}
