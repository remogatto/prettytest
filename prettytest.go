/*

Copyright (c) 2010 Andrea Fazzi

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

*/

/*

PrettyTest is a simple testing library for golang. It aims to
simplify/prettify testing in golang.

It features:

* a simple assertion vocabulary for better readability

* customizable formatters through interfaces

* before/after functions

* integrated with the go test command

* pretty and colorful output with reports

This is the skeleton of a typical prettytest test file:

    package foo

    import (
	"testing"
	"github.com/remogatto/prettytest"
    )

    // Start of setup
    type testSuite struct {
	prettytest.Suite
    }

    func TestRunner(t *testing.T) {
	prettytest.Run(
		t,
		new(testSuite),
	)
    }
    // End of setup


    // Tests start here
    func (t *testSuite) TestTrueIsTrue() {
	t.True(true)
    }

See example/example_test.go and prettytest_test.go for comprehensive
usage examples.

*/
package prettytest

import (
	"flag"
	"fmt"
	"launchpad.net/gocheck"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

const (
	STATUS_NO_ASSERTIONS = iota
	STATUS_PASS
	STATUS_FAIL
	STATUS_MUST_FAIL
	STATUS_PENDING
)

const formatTag = "\t%s\t"

var (
	testToRun = flag.String("pt.run", "", "[prettytest] regular expression that filters tests and examples to run")
	ErrorLog  []*Error
)

func green(text string) string {
	return "\033[32m" + text + "\033[0m"
}

func red(text string) string {
	return "\033[31m" + text + "\033[0m"
}

func yellow(text string) string {
	return "\033[33m" + text + "\033[0m"
}

var (
	labelFAIL         = red("F")
	labelMUSTFAIL     = green("EF")
	labelPASS         = green("OK")
	labelPENDING      = yellow("PE")
	labelNOASSERTIONS = yellow("NA")
)

type Error struct {
	Suite     *Suite
	TestFunc  *TestFunc
	Assertion *Assertion
}

type callerInfo struct {
	name, fn string
	line     int
}

func newCallerInfo(skip int) *callerInfo {
	pc, fn, line, ok := runtime.Caller(skip)
	if !ok {
		panic("An error occured while retrieving caller info!")
	}
	splits := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	callerName := splits[len(splits)-1]
	return &callerInfo{callerName, fn, line}
}

type tCatcher interface {
	setT(t *testing.T)
	suite() *Suite
	setSuiteName(name string)
	testFuncs() map[string]*TestFunc
	init()
}

type Assertion struct {
	Line         int
	Name         string
	Filename     string
	ErrorMessage string
	Passed       bool
	suite        *Suite
	testFunc     *TestFunc
}

func logError(error *Error) {
	ErrorLog = append(ErrorLog, error)
}

func (assertion *Assertion) fail() {
	assertion.Passed = false
	assertion.testFunc.Status = STATUS_FAIL
	logError(&Error{assertion.suite, assertion.testFunc, assertion})
}

type TestFunc struct {
	Name, CallerName string
	Status           int
	Assertions       []*Assertion
	suite            *Suite
	mustFail         bool
}

type Suite struct {
	T         *testing.T
	Name      string
	TestFuncs map[string]*TestFunc
}

func (s *Suite) setT(t *testing.T) { s.T = t }
func (s *Suite) init() {
	s.TestFuncs = make(map[string]*TestFunc)
}
func (s *Suite) suite() *Suite                   { return s }
func (s *Suite) setSuiteName(name string)        { s.Name = name }
func (s *Suite) testFuncs() map[string]*TestFunc { return s.TestFuncs }

func (s *Suite) appendTestFuncFromMethod(method *callerInfo) *TestFunc {
	if _, ok := s.TestFuncs[method.name]; !ok {
		s.TestFuncs[method.name] = &TestFunc{Name: method.name, Status: STATUS_PASS, suite: s}
	}
	return s.TestFuncs[method.name]
}

func (s *Suite) currentTestFunc() *TestFunc {
	callerName := newCallerInfo(3).name
	_, ok := s.TestFuncs[callerName]
	if !ok {
		s.TestFuncs[callerName] = &TestFunc{
			Name:   callerName,
			Status: STATUS_NO_ASSERTIONS,
		}
	}
	return s.TestFuncs[callerName]
}

func (testFunc *TestFunc) resetLastError() {
	if len(ErrorLog) > 0 {
		lastError := ErrorLog[len(ErrorLog)-1]
		lastError.Assertion.Passed = true
		ErrorLog = append(ErrorLog[:len(ErrorLog)-1])
		testFunc.Status = STATUS_PASS
		for i := 0; i < len(testFunc.Assertions); i++ {
			if !testFunc.Assertions[i].Passed {
				testFunc.Status = STATUS_FAIL
			}
		}
	}
}

func (testFunc *TestFunc) logError(message string) {
	assertion := &Assertion{ErrorMessage: message}
	error := &Error{testFunc.suite, testFunc, assertion}
	logError(error)
}

func (testFunc *TestFunc) appendAssertion(assertion *Assertion) *Assertion {
	testFunc.Assertions = append(testFunc.Assertions, assertion)
	return assertion
}

func (testFunc *TestFunc) status() int {
	return testFunc.Status
}

func (s *Suite) setup(errorMessage string, customMessages []string) *Assertion {
	var message string
	if len(customMessages) > 0 {
		message = strings.Join(customMessages, "\t\t\n")
	} else {
		message = errorMessage
	}
	// Retrieve the testing method
	callerInfo := newCallerInfo(3)
	assertionName := newCallerInfo(2).name
	testFunc := s.appendTestFuncFromMethod(callerInfo)
	assertion := &Assertion{
		Line:         callerInfo.line,
		Filename:     callerInfo.fn,
		Name:         assertionName,
		suite:        s,
		testFunc:     testFunc,
		ErrorMessage: message,
		Passed:       true,
	}
	testFunc.appendAssertion(assertion)
	return assertion
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

// Run runs the test suites.
func Run(t *testing.T, suites ...tCatcher) {
	run(t, new(TDDFormatter), suites...)
}

// Run runs the test suites using the given formatter.
func RunWithFormatter(t *testing.T, formatter Formatter, suites ...tCatcher) {
	run(t, formatter, suites...)
}

// Run tests. Use default formatter.
func run(t *testing.T, formatter Formatter, suites ...tCatcher) {
	var (
		beforeAllFound, afterAllFound                                                    bool
		beforeAll, afterAll, before, after                                               reflect.Value
		totalPassed, totalFailed, totalPending, totalNoAssertions, totalExpectedFailures int
	)

	flag.Parse()

	for _, s := range suites {
		beforeAll, afterAll, before, after = reflect.Value{}, reflect.Value{}, reflect.Value{}, reflect.Value{}
		s.setT(t)
		s.init()

		iType := reflect.TypeOf(s)

		s.setSuiteName(strings.Split(iType.String(), ".")[1])
		formatter.PrintSuiteInfo(s.suite())

		// search for Before and After methods
		for i := 0; i < iType.NumMethod(); i++ {
			method := iType.Method(i)
			if ok, _ := regexp.MatchString("^BeforeAll", method.Name); ok {
				if !beforeAllFound {
					beforeAll = method.Func
					beforeAllFound = true
					continue
				}
			}
			if ok, _ := regexp.MatchString("^AfterAll", method.Name); ok {
				if !afterAllFound {
					afterAll = method.Func
					afterAllFound = true
					continue
				}
			}
			if ok, _ := regexp.MatchString("^Before", method.Name); ok {
				before = method.Func
			}
			if ok, _ := regexp.MatchString("^After", method.Name); ok {
				after = method.Func
			}
		}

		if beforeAll.IsValid() {
			beforeAll.Call([]reflect.Value{reflect.ValueOf(s)})
		}

		for i := 0; i < iType.NumMethod(); i++ {
			method := iType.Method(i)
			if ok, _ := regexp.MatchString(*testToRun, method.Name); ok {
				if ok, _ := regexp.MatchString(formatter.AllowedMethodsPattern(), method.Name); ok {
					if before.IsValid() {
						before.Call([]reflect.Value{reflect.ValueOf(s)})
					}

					method.Func.Call([]reflect.Value{reflect.ValueOf(s)})

					if after.IsValid() {
						after.Call([]reflect.Value{reflect.ValueOf(s)})
					}

					testFunc, ok := s.testFuncs()[method.Name]
					if !ok {
						testFunc = &TestFunc{Name: method.Name, Status: STATUS_NO_ASSERTIONS}
					}

					if testFunc.mustFail {
						if testFunc.Status != STATUS_FAIL {
							testFunc.Status = STATUS_FAIL
							testFunc.logError("The test was expected to fail")
						} else {
							testFunc.Status = STATUS_MUST_FAIL
						}
					}

					switch testFunc.Status {
					case STATUS_PASS:
						totalPassed++
					case STATUS_FAIL:
						totalFailed++
						t.Fail()
					case STATUS_MUST_FAIL:
						totalExpectedFailures++
					case STATUS_PENDING:
						totalPending++
					case STATUS_NO_ASSERTIONS:
						totalNoAssertions++
					}
					formatter.PrintStatus(testFunc)
				}

			}

		}

		if afterAll.IsValid() {
			afterAll.Call([]reflect.Value{reflect.ValueOf(s)})
		}
	}
	formatter.PrintErrorLog(ErrorLog)
	formatter.PrintFinalReport(&FinalReport{Passed: totalPassed,
		Failed:           totalFailed,
		Pending:          totalPending,
		ExpectedFailures: totalExpectedFailures,
		NoAssertions:     totalNoAssertions,
	})
}

func init() {
	ErrorLog = make([]*Error, 0)
}
