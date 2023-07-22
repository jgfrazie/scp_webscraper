package utils

// Functions similarly to the map higher-order function in Clojure and Haskell.
// It takes a function which can be applied to each element in a slice and a slice of that type, then
// returns a slice which each element of coll has had appiledFunc invoked on.
func Map[T any, U any](appliedFunc func (T) U, coll *[]T) *[]U {
	var newColl []U = []U{}
	for _, element := range *coll {
		newColl = append(newColl, appliedFunc(element))
	}
	return &newColl
}
