// Package repository is the user data layer.
package repository

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type UserRepository interface {
	FindAllUsers() []User
	InsertUser(u User) User
}

type userRepository struct {
	rows []User
	next int
}

func New() UserRepository { return &userRepository{rows: []User{{ID: 1, Name: "Ada"}}, next: 2} }

func (r *userRepository) FindAllUsers() []User { return r.rows }

func (r *userRepository) InsertUser(u User) User {
	u.ID = r.next
	r.next++
	r.rows = append(r.rows, u)
	return u
}
