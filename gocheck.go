// +build !android

package prettytest

import (
	"fmt"
	"launchpad.net/gocheck"
)

// Check wraps gocheck.Check method.
// It verifies if the first value matches with the expected value.
// What matching means is defined by the provided checker. In case
// they do not match, an error will be logged, the test will be marked
// as failed, and the test execution will continue.  Some checkers may
// not need the expected argument (e.g. IsNil).  In either case, any
// extra arguments provided to the function will be logged next to the
// reported problem when the matching fails.  This is a handy way to
// provide problem-specific hints. (taken from gocheck doc)
func (s *Suite) Check(obtained interface{}, checker gocheck.Checker, args ...interface{}) *Assertion {
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
	return assertion
}
