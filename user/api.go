package api

import ()

type User struct {
}

func (u *User) TextMessage(params *string, _ *struct{}) error {
	return nil
}
