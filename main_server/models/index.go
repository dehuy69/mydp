package models

import mapset "github.com/deckarep/golang-set/v2"

// Generic Node type
// type Node[T any] struct {
// 	Value T
// 	Keys  mapset.Set[string]
// }

// Generic Node type
type Node struct {
	Value []byte
	Keys  mapset.Set[string]
}
