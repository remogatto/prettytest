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

PrettyTest is a simple assertion testing library for golang. It aims
to simplify/prettify testing in golang.

It features:

  * a simple assertion vocabulary for better readability
  * customizable formatters through interfaces
  * integrated with the go test command
  * pretty and colorful output with reports

See prettytest_test.go for an usage example.

*/
package prettytest

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

const (
	STATUS_PASS = iota
	STATUS_FAIL
	STATUS_PENDING
)

const formatTag = "\t%s\t"

var testToRun = flag.String("pt.run", "", "[prettytest] regular expression to select tests and examples to run")

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
	labelFAIL    = red("FAIL")
	labelPASS    = green("OK")
	labelPENDING = yellow("PENDING")
)

type callerInfo struct {
	name, fn string
	line     int
}

func newCallerInfo(skip int) *callerInfo {
	pc, fn, line, ok := runtime.Caller(skip)
	if !ok {
		panic("An error occured while retrieving caller info!")
	}
	callerName := strings.Join(strings.Split(runtime.FuncForPC(pc).Name(), ".")[1:], ".")
	return &callerInfo{callerName, fn, line}
}

type TCatcher interface {
	SetT(t *testing.T)
	GetLastStatus() byte
	GetStatus() byte
	SetStatus(status byte)
	GetInfo() *suiteInfo
	Reset()
}

type suiteInfo struct {
	assertions int
	callerName string
}

type Suite struct {
	T                  *testing.T
	Status, LastStatus byte
	callerInfo         *callerInfo
	info               map[string]*suiteInfo
}

// Formatter is the interface each formatter should implement.
type Formatter interface {
	PrintSuiteName(name string)
	PrintStatus(status byte, info *suiteInfo)
	PrintFinalReport(passed, failed, pending int)

	// AllowedMethodPattern returns a regexp for the allowed
	// method name (e.g. "^Test.*" for the TDDFormatter)
	AllowedMethodsPattern() string
}

type TDDFormatter struct{}

func (formatter *TDDFormatter) PrintSuiteName(name string) {
	fmt.Printf("\n%s:\n", name)
}

func (formatter *TDDFormatter) PrintStatus(status byte, info *suiteInfo) {
	callerName := info.callerName
	if strings.Contains(callerName, ".") {
		t := strings.Split(callerName, ".")
		callerName = t[len(t)-1]
	}

	switch status {
	case STATUS_FAIL:
		fmt.Printf(formatTag+"%-30s(%d assertion(s))\n", labelFAIL, callerName, info.assertions)
	case STATUS_PASS:
		fmt.Printf(formatTag+"%-30s(%d assertion(s))\n", labelPASS, callerName, info.assertions)
	case STATUS_PENDING:
		fmt.Printf(formatTag+"%s\n", labelPENDING, callerName)
	}
}

func (formatter *TDDFormatter) PrintFinalReport(passed, failed, pending int) {
	total := passed + failed + pending
	fmt.Printf("\n%d tests, %d passed, %d failed, %d pending\n", total, passed, failed, pending)
}

func (formatter *TDDFormatter) AllowedMethodsPattern() string {
	return "^Test.*"
}

type BDDFormatter struct {
	Description string
}

func (formatter *BDDFormatter) PrintSuiteName(name string) {
	fmt.Printf("\n%s:\n", formatter.Description)
}

func (formatter *BDDFormatter) PrintStatus(status byte, info *suiteInfo) {
	shouldText := info.callerName
	if strings.Contains(info.callerName, ".") {
		shouldText = formatter.splitString(info.callerName, ".")
	}

	switch status {
	case STATUS_FAIL:
		fmt.Printf("- %s\n", red(shouldText))
	case STATUS_PASS:
		fmt.Printf("- %s\n", green(shouldText))
	case STATUS_PENDING:
		fmt.Printf("- %s\t(Not Yet Implemented)\n", yellow(shouldText))
	}
}

func (formatter *BDDFormatter) PrintFinalReport(passed, failed, pending int) {
	total := passed + failed + pending
	fmt.Printf("\n%d examples, %d passed, %d failed, %d pending\n", total, passed, failed, pending)
}

func (formatter *BDDFormatter) AllowedMethodsPattern() string {
	return "^Should_.*"
}

func (formatter *BDDFormatter) splitString(text, sep string) (result string) {
	s := strings.Split(text, sep)

	if len(s) < 2 {
		panic("Can't use BDD formatter!")
	}

	stringWithUnderscores := s[2]
	splittedByUnderscores := strings.Split(stringWithUnderscores, "_")

	for _, v := range splittedByUnderscores {
		result += v + " "
	}
	return strings.TrimSpace(result)
}

func (s *Suite) SetT(t *testing.T)          { s.T = t }
func (s *Suite) GetLastStatus() byte        { return s.LastStatus }
func (s *Suite) GetStatus() byte            { return s.Status }
func (s *Suite) SetStatus(status byte)      { s.Status = status }
func (s *Suite) GetCallerInfo() *callerInfo { return s.callerInfo }
func (s *Suite) GetInfo() *suiteInfo        { return s.info[s.callerInfo.name] }

func (s *Suite) Reset() {
	s.info = make(map[string]*suiteInfo)
}

func (s *Suite) fail(exp, act interface{}, info *callerInfo) {
	s.T.Errorf("Expected %v but got %v -- %s:%d\n", exp, act, info.fn, info.line)
	s.Status, s.LastStatus = STATUS_FAIL, STATUS_FAIL
}

func (s *Suite) failWithCustomMsg(msg string, info *callerInfo) {
	s.T.Errorf("%s -- %s:%d\n", msg, info.fn, info.line)
	s.Status, s.LastStatus = STATUS_FAIL, STATUS_FAIL
}

func (s *Suite) setup() {
	s.LastStatus = STATUS_PASS
	s.callerInfo = newCallerInfo(3)
	if _, present := s.info[s.callerInfo.name]; !present {
		s.info[s.callerInfo.name] = new(suiteInfo)

	}
	s.info[s.callerInfo.name].callerName = s.callerInfo.name
	s.info[s.callerInfo.name].assertions++
}

// Equal asserts that the expected value equals the actual value.
func (s *Suite) Equal(exp, act interface{}) {
	s.setup()
	if exp != act {
		s.fail(exp, act, s.callerInfo)
	}
}

// NotEqual asserts that the expected value is not equal to the actual
// value.
func (s *Suite) NotEqual(exp, act interface{}) {
	s.setup()
	if exp == act {
		s.fail(exp, act, s.callerInfo)
	}
}

// True asserts that the value is true.
func (s *Suite) True(value bool) {
	s.setup()
	if !value {
		s.fail("true", "false", s.callerInfo)
	}
}

// False asserts that the value is false.
func (s *Suite) False(value bool) {
	s.setup()
	if value {
		s.fail("false", "true", s.callerInfo)
	}
}

// Path asserts that the given path exists.
func (s *Suite) Path(path string) {
	s.setup()
	if _, err := os.Stat(path); err != nil {
		s.failWithCustomMsg(fmt.Sprintf("Path %s doesn't exist", path), s.callerInfo)
	}
}

// Nil asserts that the value is nil.
func (s *Suite) Nil(value interface{}) {
	s.setup()
	if value != nil {
		s.failWithCustomMsg(fmt.Sprintf("Expected nil but got %s", value), s.callerInfo)
	}
}

// NotNil asserts that the value is not nil.
func (s *Suite) NotNil(value interface{}) {
	s.setup()
	if value == nil {
		s.failWithCustomMsg(fmt.Sprintf("Expected not nil value but got %s", value), s.callerInfo)
	}
}

// Pending marks the test function as pending.
func (s *Suite) Pending() {
	s.setup()
	s.Status = STATUS_PENDING
}

// Failed checks if the last assertion has failed.
func (s *Suite) Failed() bool {
	return s.LastStatus == STATUS_FAIL
}

// FailedTest checks if the test function has failed.
func (s *Suite) FailedTest() bool {
	return s.Status == STATUS_FAIL
}

// Run runs the test suites.
func Run(t *testing.T, suites ...TCatcher) {
	run(t, new(TDDFormatter), suites...)
}

// Run runs the test suites using the given formatter.
func RunWithFormatter(t *testing.T, formatter Formatter, suites ...TCatcher) {
	run(t, formatter, suites...)
}

// Run tests. Use default formatter.
func run(t *testing.T, formatter Formatter, suites ...TCatcher) {
	var (
		beforeAllFound, afterAllFound          bool
		beforeAll, afterAll, before, after     reflect.Value
		totalPassed, totalFailed, totalPending int
	)

	flag.Parse()

	for _, s := range suites {
		beforeAll, afterAll, before, after = reflect.Value{}, reflect.Value{}, reflect.Value{}, reflect.Value{}
		s.SetT(t)
		s.Reset()

		iType := reflect.TypeOf(s)

		formatter.PrintSuiteName(strings.Split(iType.String(), ".")[1])

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

					s.SetStatus(STATUS_PASS)

					if before.IsValid() {
						before.Call([]reflect.Value{reflect.ValueOf(s)})
					}
					method.Func.Call([]reflect.Value{reflect.ValueOf(s)})

					if after.IsValid() {
						after.Call([]reflect.Value{reflect.ValueOf(s)})
					}

					switch s.GetStatus() {
					case STATUS_PASS:
						totalPassed++
					case STATUS_FAIL:
						totalFailed++
					case STATUS_PENDING:
						totalPending++
					}

					formatter.PrintStatus(s.GetStatus(), s.GetInfo())
				}
			}

		}

		if afterAll.IsValid() {
			afterAll.Call([]reflect.Value{reflect.ValueOf(s)})
		}
	}

	formatter.PrintFinalReport(totalPassed, totalFailed, totalPending)
}
