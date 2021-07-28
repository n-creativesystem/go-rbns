package tests

import "testing"

type Case struct {
	Name string
	Fn   func(t *testing.T)
}

type Cases []Case

func (c Cases) Run(t *testing.T) {
	for _, tt := range c {
		t.Run(tt.Name, tt.Fn)
	}
}
