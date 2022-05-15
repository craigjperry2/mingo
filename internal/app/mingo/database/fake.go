package database

import (
	"sync"

	"github.com/craigjperry2/mingo/internal/app/mingo"
)

// Fake DB access layer for this iteration
type DbNothingBurger struct {
	mu       sync.Mutex
	rows     []mingo.Person
	sequence int
}

func NewDatabase() *DbNothingBurger {
	return &DbNothingBurger{sync.Mutex{}, []mingo.Person{}, 0}
}

func (db *DbNothingBurger) NextId() int {
	db.sequence++
	return db.sequence
}

func (db *DbNothingBurger) GetAll(offset int, limit int) ([]mingo.Person, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	// Because of my HTMX UI impl, i can't get away with simple offset/limit pagination, so using cursor style instead
	var result []mingo.Person
	// result := make([]Person, limit)
	for i, j := 0, 0; i < len(db.rows) && j < limit; i++ {
		if db.rows[i].Id <= offset { // assumes db idx dont start at 0
			continue
		}
		result = append(result, db.rows[i])
		j++
	}
	return result, nil
}

func (db *DbNothingBurger) Get(id int) (mingo.Person, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	// Because of my HTMX UI impl, i can't get away with simple offset/limit pagination, so using cursor style instead
	var result mingo.Person
	// result := make([]Person, limit)
	for i := 0; i < len(db.rows); i++ {
		if db.rows[i].Id == id { // assumes db idx dont start at 0
			result = db.rows[i]
		}
	}
	return result, nil
}

func (db *DbNothingBurger) Update(id int, name string, location string) (mingo.Person, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	p := mingo.Person{Id: id, Name: name, Location: location}
	for i := 0; i < len(db.rows); i++ {
		if db.rows[i].Id == id {
			db.rows[i] = p
		}
	}
	return p, nil
}

// Invariant: inserts must be append only with monotonic key, because i'm using cursor paginaiton in GetAll
func (db *DbNothingBurger) Insert(name string, location string) (mingo.Person, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	p := mingo.Person{Id: db.NextId(), Name: name, Location: location}
	db.rows = append(db.rows, p) // Defensive copy not needed, pass by value in Go
	return p, nil
}

func (db *DbNothingBurger) Delete(id int) (mingo.Person, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	var old mingo.Person
	for i, p := range db.rows {
		if p.Id == id {
			old = p
			db.rows = append(db.rows[:i], db.rows[i+1:]...)
			break
		}
	}
	return old, nil
}
