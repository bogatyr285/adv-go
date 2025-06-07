package hw9

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Person struct {
	Name    string `properties:"name"`
	Address string `properties:"address,omitempty"`
	Age     int    `properties:"age"`
	Married bool   `properties:"married"`
}

func Serialize(obj interface{}) string {
	v := reflect.ValueOf(obj)
	t := v.Type()

	// estimate row size:
	//   10 chars key
	// + 15 chars value
	// + 2 chars for = and \n
	// = 27
	estimatedSize := t.NumField() * 30
	var builder strings.Builder
	builder.Grow(estimatedSize)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("properties")
		if tag == "" {
			continue
		}

		parts := strings.Split(tag, ",")
		name := parts[0]
		omitempty := len(parts) > 1 && parts[1] == "omitempty"

		fieldValue := v.Field(i)
		zero := reflect.Zero(fieldValue.Type()).Interface()

		if omitempty && reflect.DeepEqual(fieldValue.Interface(), zero) {
			continue
		}

		builder.WriteString(name)
		builder.WriteByte('=')

		switch fieldValue.Kind() {
		case reflect.String:
			builder.WriteString(fieldValue.String())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			builder.WriteString(strconv.FormatInt(fieldValue.Int(), 10))
		case reflect.Bool:
			if fieldValue.Bool() {
				builder.WriteString("true")
			} else {
				builder.WriteString("false")
			}
		}

		builder.WriteByte('\n')
	}

	return strings.TrimSuffix(builder.String(), "\n")
}

func TestSerialization(t *testing.T) {
	tests := map[string]struct {
		person Person
		result string
	}{
		"test case with empty fields": {
			result: "name=\nage=0\nmarried=false",
		},
		"test case with fields": {
			person: Person{
				Name:    "John Doe",
				Age:     30,
				Married: true,
			},
			result: "name=John Doe\nage=30\nmarried=true",
		},
		"test case with omitempty field": {
			person: Person{
				Name:    "John Doe",
				Age:     30,
				Married: true,
				Address: "Paris",
			},
			result: "name=John Doe\naddress=Paris\nage=30\nmarried=true",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := Serialize(test.person)
			assert.Equal(t, test.result, result)
		})
	}
}
