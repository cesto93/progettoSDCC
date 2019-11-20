package utility

import (
	"errors"
	"strings"
)

type ArrayFlags []string

func (i *ArrayFlags) String() string {
    return "my string representation"
}

func (i *ArrayFlags) Set(value string) error {
	if len(*i) > 0 {
		return errors.New("interval flag already set")
	}
	for _, dt := range strings.Split(value, ",") {
		*i = append(*i, dt)
	}
	return nil
}