package e

import "fmt"

func Wrap(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, err)
}

func WrapErr(err error, msg string) error {
	if err == nil {
		return nil
	}
	return Wrap(err, msg)
}
