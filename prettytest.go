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
	"reflect"
	"regexp"
	"runtime"
	"strings"
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
	ErrorLog          []*Error
	labelFAIL         = red("F")
	labelMUSTFAIL     = green("EF")
	labelPASS         = green("OK")
	labelPENDING      = yellow("PE")
	labelNOASSERTIONS = yellow("NA")
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
	return &callerInfo{splits[len(splits)-1], fn, line}
}

type tCatcher interface {
	setT(t T)
	suite() *Suite
	setPackageName(name string)
	setSuiteName(name string)
	testFuncs() map[string]*TestFunc
	init()
}

func logError(error *Error) {
	ErrorLog = append(ErrorLog, error)
}

type TestFunc struct {
	Name, CallerName string
	Status           int
	Assertions       []*Assertion
	suite            *Suite
	mustFail         bool
}

type T interface {
	Fail()
}

type Suite struct {
	T             T
	Package, Name string
	TestFuncs     map[string]*TestFunc
}

func (s *Suite) setT(t T)                        { s.T = t }
func (s *Suite) init()                           { s.TestFuncs = make(map[string]*TestFunc) }
func (s *Suite) suite() *Suite                   { return s }
func (s *Suite) setPackageName(name string)      { s.Package = name }
func (s *Suite) setSuiteName(name string)        { s.Name = name }
func (s *Suite) testFuncs() map[string]*TestFunc { return s.TestFuncs }

func (s *Suite) appendTestFuncFromMethod(method *callerInfo) *TestFunc {
	name := method.name
	if _, ok := s.TestFuncs[name]; !ok {
		s.TestFuncs[name] = &TestFunc{
			Name:   name,
			Status: STATUS_PASS,
			suite:  s,
		}
	}
	return s.TestFuncs[name]
}

func (s *Suite) currentTestFunc() *TestFunc {
	callerName := newCallerInfo(3).name
	if _, ok := s.TestFuncs[callerName]; !ok {
		s.TestFuncs[callerName] = &TestFunc{
			Name:   callerName,
			Status: STATUS_NO_ASSERTIONS,
		}
	}
	return s.TestFuncs[callerName]
}

func (testFunc *TestFunc) resetLastError() {
	if len(ErrorLog) > 0 {
		ErrorLog[len(ErrorLog)-1].Assertion.Passed = true
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

// Run runs the test suites.
func Run(t T, suites ...tCatcher) {
	run(t, new(TDDFormatter), suites...)
}

// Run runs the test suites using the given formatter.
func RunWithFormatter(t T, formatter Formatter, suites ...tCatcher) {
	run(t, formatter, suites...)
}

// Run tests. Use default formatter.
func run(t T, formatter Formatter, suites ...tCatcher) {
	var (
		beforeAllFound, afterAllFound                                                    bool
		beforeAll, afterAll, before, after                                               reflect.Value
		totalPassed, totalFailed, totalPending, totalNoAssertions, totalExpectedFailures int
	)

	ErrorLog = make([]*Error, 0)
	//	flag.Parse()

	for _, s := range suites {
		beforeAll, afterAll, before, after = reflect.Value{}, reflect.Value{}, reflect.Value{}, reflect.Value{}
		s.setT(t)
		s.init()

		iType := reflect.TypeOf(s)
		splits := strings.Split(iType.String(), ".")
		s.setPackageName(splits[0][1:])
		s.setSuiteName(splits[1])
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
			if filterMethod(method.Name) {
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
}
