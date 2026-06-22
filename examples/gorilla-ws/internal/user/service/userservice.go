// Package service holds user business logic.
package service

import "archview-example-gorilla-ws/internal/user/repository"

type UserService interface {
	ListUsers() []repository.User
	CreateUser(name string) repository.User
}

type userService struct{ repo repository.UserRepository }

func New(repo repository.UserRepository) UserService { return &userService{repo: repo} }

func (s *userService) ListUsers() []repository.User { return s.repo.FindAllUsers() }
func (s *userService) CreateUser(name string) repository.User {
	return s.repo.InsertUser(repository.User{Name: name})
}
