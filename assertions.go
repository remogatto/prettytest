package prettytest

import (
	"fmt"
	"launchpad.net/gocheck"
	"os"
	"reflect"
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

func (s *Suite) Check(obtained interface{}, checker gocheck.Checker, args ...interface{}) bool {
	assertion := s.setup("", []string{})
	checkerInfo := checker.Info()
	params := make([]interface{}, len(args)+1)
	params[0] = obtained
	copy(params[1:], args)
	result, _ := checker.Check(params, []string{})
	if !result {
		errorMsg := fmt.Sprintf("%s checker failed: ", checkerInfo.Name)
		for i, param := range checkerInfo.Params {
			errorMsg += fmt.Sprintf("%s %v ", param, params[i])
		}
		assertion.ErrorMessage = errorMsg
		assertion.fail()
	} else {
		assertion.Passed = true
	}
	return assertion.Passed
}

func (s *Suite) Not(result bool, messages ...string) bool {
	assertion := s.setup(fmt.Sprintf("Expected assertion to fail"), messages)
	if result {
		assertion.fail()
	} else {
		assertion.testFunc.resetLastError()
	}
	return assertion.Passed
}

// Equal asserts that the expected value equals the actual value.
func (s *Suite) Equal(exp, act interface{}, messages ...string) bool {
	assertion := s.setup(fmt.Sprintf("Expected %v to be equal to %v", act, exp), messages)
	if exp != act {
		assertion.fail()
	}
	return assertion.Passed
}

// True asserts that the value is true.
func (s *Suite) True(value bool, messages ...string) bool {
	assertion := s.setup(fmt.Sprintf("Expected value to be true"), messages)
	if !value {
		assertion.fail()
	}
	return assertion.Passed
}

// Path asserts that the given path exists.
func (s *Suite) Path(path string, messages ...string) bool {
	assertion := s.setup(fmt.Sprintf("Path %s doesn't exist", path), messages)
	if _, err := os.Stat(path); err != nil {
		assertion.fail()
	}
	return assertion.Passed
}

// Nil asserts that the value is nil.
func (s *Suite) Nil(value interface{}, messages ...string) bool {
	assertion := s.setup(fmt.Sprintf("Value %v is not nil", value), messages)
	if value == nil {
		return true
	}
	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		if !v.IsNil() {
			assertion.fail()
		}
	}
	return assertion.Passed
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

func (s *Suite) MustFail() {
	s.currentTestFunc().mustFail = true
}

// // Fail marks the test as failed
// func (s *Suite) Fail() {
// 	s.Status.Code = STATUS_FAIL
// }

// // Failed checks if the last assertion has failed.
// func (s *Suite) Failed() bool {
// 	return s.Status.LastCode == STATUS_FAIL
// }

// // FailedTest checks if the test function has failed.
// func (s *Suite) FailedTest() bool {
// 	return s.Status.Code == STATUS_FAIL
// }
