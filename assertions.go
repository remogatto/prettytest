package prettytest

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

type Assertion struct {
	Line         int
	Name         string
	Filename     string
	ErrorMessage string
	Passed       bool
	suite        *Suite
	testFunc     *TestFunc
}

func (assertion *Assertion) fail() {
	assertion.Passed = false
	assertion.testFunc.Status = STATUS_FAIL
	logError(&Error{assertion.suite, assertion.testFunc, assertion})
}

// Not asserts the given assertion is false.
func (s *Suite) Not(result *Assertion, messages ...string) *Assertion {
	assertion := s.setup(fmt.Sprintf("Expected assertion to fail"), messages)
	if result.Passed {
		assertion.fail()
	} else {
		result.Passed = true
		assertion.testFunc.resetLastError()
	}
	return assertion
}

// Not asserts the given assertion is false.
func (s *Suite) False(value bool, messages ...string) *Assertion {
	assertion := s.setup(fmt.Sprintf("Expected value to be false"), messages)
	if value {
		assertion.fail()
	}
	return assertion
}

// Equal asserts that the expected value equals the actual value.
func (s *Suite) Equal(exp, act interface{}, messages ...string) *Assertion {
	actType := reflect.TypeOf(act)
	expType := reflect.TypeOf(exp)
	assertion := s.setup(fmt.Sprintf("Expected %v[%s] to be equal to %v[%s]", act, actType, exp, expType), messages)
	if exp != act {
		assertion.fail()
	}
	return assertion
}

// Contain asserts that the actual value contains the expected value.
func (s *Suite) Contain(exp, act interface{}, messages ...string) *Assertion {
	actType := reflect.TypeOf(act)
	expType := reflect.TypeOf(exp)
	assertion := s.setup(fmt.Sprintf("Expected %v[%s] to be contained in %v[%s]", act, actType, exp, expType), messages)
	if !strings.Contains(act.(string), exp.(string)) {
		assertion.fail()
	}
	return assertion
}

// True asserts that the value is true.
func (s *Suite) True(value bool, messages ...string) *Assertion {
	assertion := s.setup(fmt.Sprintf("Expected value to be true"), messages)
	if !value {
		assertion.fail()
	}
	return assertion
}

// Path asserts that the given path exists.
func (s *Suite) Path(path string, messages ...string) *Assertion {
	assertion := s.setup(fmt.Sprintf("Path %s doesn't exist", path), messages)
	if _, err := os.Stat(path); err != nil {
		assertion.fail()
	}
	return assertion
}

// Nil asserts that the value is nil.
func (s *Suite) Nil(value interface{}, messages ...string) *Assertion {
	assertion := s.setup(fmt.Sprintf("Value %v is not nil", value), messages)
	if value == nil {
		return assertion
	}
	val := reflect.ValueOf(value)
	val.Kind()
	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		if !v.IsNil() {
			assertion.fail()
		}
	}
	return assertion
}

// Error logs an error and marks the test function as failed.
func (s *Suite) Error(args ...interface{}) {
	assertion := s.setup("", []string{})
	assertion.testFunc.Status = STATUS_FAIL
	assertion.ErrorMessage = fmt.Sprint(args...)
	assertion.fail()
}

// Pending marks the test function as pending.
func (s *Suite) Pending() {
	s.currentTestFunc().Status = STATUS_PENDING
}

// MustFail marks the current test function as an expected failure.
func (s *Suite) MustFail() {
	s.currentTestFunc().mustFail = true
}

// Failed checks if the test function has failed.
func (s *Suite) Failed() bool {
	return s.currentTestFunc().Status == STATUS_FAIL
}
