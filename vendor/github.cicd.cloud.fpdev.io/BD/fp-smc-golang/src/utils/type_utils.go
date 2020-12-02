package utils

import (
	"reflect"
)

type AnonymousType reflect.Value

func ToAnonymousType(obj interface{}) AnonymousType {
	return AnonymousType(reflect.ValueOf(obj))
}

func (a AnonymousType) IsA(typeToAssert reflect.Kind) bool {
	return typeToAssert == reflect.Value(a).Kind()
}
