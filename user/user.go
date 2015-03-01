package main

import ()

type User struct {
}

func newUser() *User {
	return &User{}
}

func (u *User) TextMessage(params *string, _ *struct{}) error {
	return nil
}
