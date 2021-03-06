// Package util contains generic helper functions
package util

import (
	"flag"
	"os"
	"strings"
)

// FlagsFromEnv override flag values with values fom environment variables if they are present
func FlagsFromEnv() {
	flag.VisitAll(func(f *flag.Flag) {
		env := strings.ToUpper(strings.Replace(f.Name, "-", "_", -1))
		if val, ok := os.LookupEnv(env); ok {
			f.Value.Set(val)
		}
	})
}

// ArrayFlag allows to set multiple similar flags into an array of strings
type ArrayFlag []string

// String gets a string representation of value
func (i *ArrayFlag) String() string {
	return strings.Join(*i, ",")
}

// Set a value
func (i *ArrayFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// StrPtr returns a pointer to string from a string
func StrPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to int from int
func IntPtr(i int) *int {
	return &i
}
