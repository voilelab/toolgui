package tcutil

import (
	"crypto/md5"
	"fmt"
	"math/rand"
)

// HashedID returns a hashed ID based on the component name and the bytes.
func HashedID(componentName string, bs []byte) string {
	return fmt.Sprintf("%s_%x", componentName, md5.Sum(bs))
}

// NormalID returns a normal ID based on the component name and the label.
func NormalID(componentName, label string) string {
	return fmt.Sprintf("%s_%s", componentName, label)
}

// RandID returns a random ID based on the component name.
func RandID(componentName string) string {
	return fmt.Sprintf("%s_%d", componentName, rand.Int())
}
