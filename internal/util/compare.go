package util

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
)

// CompareStructFields compares two named fields from two struct objects
// with flexible type conversion and detailed logging
func CompareStructFields(obj1, obj2 any, fieldName1, fieldName2 string) bool {
	val1 := reflect.ValueOf(obj1)
	val2 := reflect.ValueOf(obj2)

	// Handle pointers
	if val1.Kind() == reflect.Ptr {
		val1 = val1.Elem()
	}
	if val2.Kind() == reflect.Ptr {
		val2 = val2.Elem()
	}

	// Ensure both values are structs
	if val1.Kind() != reflect.Struct || val2.Kind() != reflect.Struct {
		log.Printf("Error: Both objects must be structs, got %v and %v", val1.Kind(), val2.Kind())
		return false
	}

	// Get field values
	field1 := val1.FieldByName(fieldName1)
	field2 := val2.FieldByName(fieldName2)

	// Check if fields exist
	if !field1.IsValid() {
		log.Printf("Error: Field '%s' not found in first struct (%T)", fieldName1, obj1)
		return false
	}
	if !field2.IsValid() {
		log.Printf("Error: Field '%s' not found in second struct (%T)", fieldName2, obj2)
		return false
	}

	// Get field types for logging
	type1 := field1.Type()
	type2 := field2.Type()

	// Try to compare with type conversion
	equal, err := compareWithTypeConversion(field1, field2)
	if err != nil {
		log.Printf("Error comparing fields: %v", err)
		return false
	}

	if !equal {
		log.Printf("Field comparison failed: %s.%s (%v: %v) != %s.%s (%v: %v)",
			reflect.TypeOf(obj1).Name(), fieldName1, type1, field1.Interface(),
			reflect.TypeOf(obj2).Name(), fieldName2, type2, field2.Interface())
	}

	return equal
}

// compareWithTypeConversion handles flexible type comparison
func compareWithTypeConversion(val1, val2 reflect.Value) (bool, error) {
	// If types are exactly the same, do direct comparison
	if val1.Type() == val2.Type() {
		return reflect.DeepEqual(val1.Interface(), val2.Interface()), nil
	}

	// Handle numeric type conversions
	if isNumeric(val1.Kind()) && isNumeric(val2.Kind()) {
		return compareNumeric(val1, val2)
	}

	// Handle string conversions
	if val1.Kind() == reflect.String || val2.Kind() == reflect.String {
		return compareAsStrings(val1, val2), nil
	}

	// Handle bool conversions
	if val1.Kind() == reflect.Bool || val2.Kind() == reflect.Bool {
		return compareAsBools(val1, val2)
	}

	// Try interface comparison as last resort
	return reflect.DeepEqual(val1.Interface(), val2.Interface()), nil
}

// isNumeric checks if a kind represents a numeric type
func isNumeric(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// compareNumeric compares two numeric values with type conversion
func compareNumeric(val1, val2 reflect.Value) (bool, error) {
	// Convert both to float64 for comparison
	f1, err := toFloat64(val1)
	if err != nil {
		return false, fmt.Errorf("cannot convert %v to float64: %v", val1.Interface(), err)
	}

	f2, err := toFloat64(val2)
	if err != nil {
		return false, fmt.Errorf("cannot convert %v to float64: %v", val2.Interface(), err)
	}

	return f1 == f2, nil
}

// toFloat64 converts various numeric types to float64
func toFloat64(val reflect.Value) (float64, error) {
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(val.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(val.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return val.Float(), nil
	default:
		return 0, fmt.Errorf("unsupported type for numeric conversion: %v", val.Kind())
	}
}

// compareAsStrings converts both values to strings and compares
func compareAsStrings(val1, val2 reflect.Value) bool {
	str1 := toString(val1)
	str2 := toString(val2)
	return str1 == str2
}

// toString converts a value to string
func toString(val reflect.Value) string {
	switch val.Kind() {
	case reflect.String:
		return val.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(val.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'g', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(val.Bool())
	default:
		return fmt.Sprintf("%v", val.Interface())
	}
}

// compareAsBools converts both values to bools and compares
func compareAsBools(val1, val2 reflect.Value) (bool, error) {
	b1, err := toBool(val1)
	if err != nil {
		return false, fmt.Errorf("cannot convert %v to bool: %v", val1.Interface(), err)
	}

	b2, err := toBool(val2)
	if err != nil {
		return false, fmt.Errorf("cannot convert %v to bool: %v", val2.Interface(), err)
	}

	return b1 == b2, nil
}

// toBool converts a value to bool
func toBool(val reflect.Value) (bool, error) {
	switch val.Kind() {
	case reflect.Bool:
		return val.Bool(), nil
	case reflect.String:
		return strconv.ParseBool(val.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() != 0, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() != 0, nil
	case reflect.Float32, reflect.Float64:
		return val.Float() != 0, nil
	default:
		return false, fmt.Errorf("unsupported type for bool conversion: %v", val.Kind())
	}
}
