package magelib

import "github.com/magefile/mage/mg"

func FatalError(err error) error {
	return mg.Fatal(1, err)
}

func Fatal(args ...interface{}) error {
	return mg.Fatal(1, args...)
}

func Fatalf(format string, args ...interface{}) error {
	return mg.Fatalf(1, format, args...)
}
