package models

import (
	"fmt"
	"reflect"
)

// UserType represents the user with email
type UserType struct {
	SimpleType
	value interface{}
}

// ConvertToModel responsible for converting input value into a valid database representable user object
func (fieldType UserType) ConvertToModel(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	valueType := reflect.TypeOf(value)
	if valueType.Kind() != reflect.String {
		return nil, fmt.Errorf("value %v should be %s, but is %s", value, "string", valueType.Name())
	}
	return value, nil
}

// ConvertFromModel converts DB value into human readable format
func (fieldType UserType) ConvertFromModel(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	valueType := reflect.TypeOf(value)
	if valueType.Kind() != reflect.String {
		return nil, fmt.Errorf("value %v should be %s, but is %s", value, "string", valueType.Name())
	}
	return value, nil
}
