// Package service holds user business logic.
package service

import "archview-example-fiber/internal/user/repository"

// UserService is the user use-case boundary (single impl: userService).
type UserService interface {
	ListUsers() []repository.User
	GetUser(id int) (repository.User, bool)
	CreateUser(name string) repository.User
}

type userService struct {
	repo repository.UserRepository
}

// New wires a UserService over a UserRepository.
func New(repo repository.UserRepository) UserService { return &userService{repo: repo} }

func (s *userService) ListUsers() []repository.User { return s.repo.FindAllUsers() }

func (s *userService) GetUser(id int) (repository.User, bool) { return s.repo.FindUserByID(id) }

func (s *userService) CreateUser(name string) repository.User {
	return s.repo.InsertUser(repository.User{Name: name})
}
