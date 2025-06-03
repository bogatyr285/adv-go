package main

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v hw8_test.go

type MultiError struct {
	errors []error
}

func (e *MultiError) Error() string {
	if e == nil || len(e.errors) == 0 {
		return ""
	}

	var sb strings.Builder
	// pre allocate buffer. ~ 20 bytes to error
	sb.Grow(20 * len(e.errors))

	sb.WriteString(strconv.Itoa(len(e.errors)))
	sb.WriteString(" ")
	if len(e.errors) == 1 {
		sb.WriteString("error")
	} else {
		sb.WriteString("errors")
	}
	sb.WriteString(" occured:\n")

	for _, err := range e.errors {
		if err != nil {
			sb.WriteString("\t* ")
			sb.WriteString(err.Error())

		}
	}
	sb.WriteString("\n")

	return sb.String()
}
func Append(err error, errs ...error) *MultiError {
	if err == nil && len(errs) == 0 {
		return nil
	}

	var me *MultiError
	if err != nil {
		var existing *MultiError
		if errors.As(err, &existing) {
			me = existing
		} else {
			me = &MultiError{errors: make([]error, 0, len(errs)+1)}
			me.errors = append(me.errors, err)
		}
	} else {
		me = &MultiError{errors: make([]error, 0, len(errs))}
	}

	for _, e := range errs {
		if e == nil {
			continue
		}

		if multiError, ok := e.(*MultiError); ok {
			me.errors = append(me.errors, multiError.errors...)
		} else {
			me.errors = append(me.errors, e)
		}
	}

	if len(me.errors) == 0 {
		return nil
	}
	return me
}

func TestMultiError(t *testing.T) {
	var err error
	err = Append(err, errors.New("error 1"))
	err = Append(err, errors.New("error 2"))

	expectedMessage := "2 errors occured:\n\t* error 1\t* error 2\n"
	assert.EqualError(t, err, expectedMessage)
}
