package main

import ()

type Interface interface {
	TextMessage(params *string, _ *struct{}) error
}
