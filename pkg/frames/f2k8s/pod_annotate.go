package f2k8s

import (
	"errors"

	frame2 "github.com/hash-d/frame2/pkg"
)

type PodAnnotate struct {
	frame2.Step
}

func (pa PodAnnotate) Run() error {
	return errors.New("not implemented")
}
