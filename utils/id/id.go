package id

import "github.com/google/uuid"

type Generator interface {
	ID() string
}

type GeneratorFunc func() string

func (f GeneratorFunc) ID() string {
	return f()
}

func UUID() string {
	return uuid.New().String()
}
