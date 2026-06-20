// Package repository is the user data layer.
package repository

// User is a demo user record.
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// UserRepository abstracts user persistence (single impl: userRepository).
type UserRepository interface {
	FindAllUsers() []User
	FindUserByID(id int) (User, bool)
	InsertUser(u User) User
}

type userRepository struct {
	rows []User
	next int
}

// New returns an in-memory UserRepository.
func New() UserRepository {
	return &userRepository{
		rows: []User{{ID: 1, Name: "Ada"}, {ID: 2, Name: "Linus"}},
		next: 3,
	}
}

func (r *userRepository) FindAllUsers() []User { return r.rows }

func (r *userRepository) FindUserByID(id int) (User, bool) {
	for _, u := range r.rows {
		if u.ID == id {
			return u, true
		}
	}
	return User{}, false
}

func (r *userRepository) InsertUser(u User) User {
	u.ID = r.next
	r.next++
	r.rows = append(r.rows, u)
	return u
}
