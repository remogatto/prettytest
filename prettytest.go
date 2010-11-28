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
* colorful output

*/

package prettytest

import (
	"testing"
	"runtime"
	"reflect"
	"regexp"
	"strings"
	"os"
	"fmt"
)

const (
	STATUS_PASS = iota
	STATUS_FAIL
	STATUS_PENDING
)

const formatTag = "\t%s\t"

func green(text string) string {
	return fmt.Sprintf("\033[32m%s\033[0m", text)
}

func red(text string) string {
	return fmt.Sprintf("\033[31m%s\033[0m", text)
}

func yellow(text string) string {
	return fmt.Sprintf("\033[33m%s\033[0m", text)
}

var (
	labelFAIL    = red("FAIL")
	labelPASS    = green("OK")
	labelPENDING = yellow("PENDING")
	labelDRY     = yellow("DRY")
)

type callerInfo struct {
	name, fn string
	line     int
}

type T struct {
	T                  *testing.T
	Status, LastStatus byte
	Dry                bool

	callerInfo *callerInfo
}

func newCallerInfo(skip int) *callerInfo {
	pc, fn, line, ok := runtime.Caller(skip)
	if !ok {
		panic("An error occured while retrieving caller info!")
	}
	callerName := runtime.FuncForPC(pc).Name()
	return &callerInfo{callerName, fn, line}
}

// Formatters
type Formatter interface {
	PrintStatus(status byte, callerInfo *callerInfo)
	PrintDry(callerInfo *callerInfo)
}

type TDDFormatter struct{}

func (formatter *TDDFormatter) PrintStatus(status byte, callerInfo *callerInfo) {
	switch status {
	case STATUS_FAIL:
		fmt.Printf(formatTag+"%s\n", labelFAIL, callerInfo.name)
	case STATUS_PASS:
		fmt.Printf(formatTag+"%s\n", labelPASS, callerInfo.name)
	case STATUS_PENDING:
		fmt.Printf(formatTag+"%s\n", labelPENDING, callerInfo.name)

	}
}

func (formatter *TDDFormatter) PrintDry(callerInfo *callerInfo) {
	fmt.Printf(formatTag+"%s\n", labelDRY, callerInfo.name)
}

type BDDFormatter struct{}

func (formatter *BDDFormatter) PrintStatus(status byte, callerInfo *callerInfo) {
	var shouldText string

	s := strings.Split(callerInfo.name, ".", -1)
	callerName := s[1]
	splittedByUnderscores := strings.Split(callerName, "_", -1)

	for _, v := range splittedByUnderscores {
		shouldText += v + " "
	}

	shouldText = strings.TrimSpace(shouldText)

	switch status {
	case STATUS_FAIL:
		fmt.Printf("- %s\n", red(shouldText))
	case STATUS_PASS:
		fmt.Printf("- %s\n", green(shouldText))
	case STATUS_PENDING:
		fmt.Printf("- %s\t(Not Yet Implemented)\n", yellow(shouldText))
	}
}

func (formatter *BDDFormatter) PrintDry(callerInfo *callerInfo) {
	fmt.Printf(formatTag+"%s\n", labelDRY, callerInfo.name)
}

func (assertion *T) fail(exp, act interface{}, info *callerInfo) {
	if !assertion.Dry {
		assertion.T.Errorf("Expected %s but got %s -- %s:%d\n", exp, act, info.fn, info.line)
	}
	assertion.Status, assertion.LastStatus = STATUS_FAIL, STATUS_FAIL
}

func (assertion *T) failWithCustomMsg(msg string, info *callerInfo) {
	if !assertion.Dry {
		assertion.T.Errorf("%s -- %s:%d\n", msg, info.fn, info.line)
	}
	assertion.Status, assertion.LastStatus = STATUS_FAIL, STATUS_FAIL
}

func (assertion *T) setup() {
	assertion.LastStatus = STATUS_PASS
	assertion.callerInfo = newCallerInfo(3)
}

// Assert that the expected value equals the actual value. Return true
// on success.
func (assertion *T) Equal(exp, act interface{}) {
	assertion.setup()
	if exp != act {
		assertion.fail(exp, act, assertion.callerInfo)
	}
}

// Assert that the value is true.
func (assertion *T) True(value bool) {
	assertion.setup()
	if !value {
		assertion.fail("true", "false", assertion.callerInfo)
	}
}

// Assert that the value is false.
func (assertion *T) False(value bool) {
	assertion.setup()
	if value {
		assertion.fail("false", "true", assertion.callerInfo)
	}
}

// Assert that the given path exists.
func (assertion *T) Path(path string) {
	assertion.setup()
	if _, err := os.Stat(path); err != nil {
		assertion.failWithCustomMsg(fmt.Sprintf("Path %s doesn't exist", path), assertion.callerInfo)
	}
}

// Assert that the value is nil.
func (assertion *T) Nil(value interface{}) {
	assertion.setup()
	if value != nil {
		assertion.failWithCustomMsg(fmt.Sprintf("Expected nil but got %s", value), assertion.callerInfo)
	}
}

// Assert that the value is not nil.
func (assertion *T) NotNil(value interface{}) {
	assertion.setup()
	if value == nil {
		assertion.failWithCustomMsg(fmt.Sprintf("Expected not nil value but got %s", value), assertion.callerInfo)
	}
}

// Mark the test function as pending.
func (assertion *T) Pending() {
	assertion.setup()
	assertion.Status = STATUS_PENDING
}

// Check if the last assertion has failed.
func (assertion *T) Failed() bool {
	return assertion.LastStatus == STATUS_FAIL
}

// Check if the test function has failed.
func (assertion *T) TestFailed() bool {
	return assertion.Status == STATUS_FAIL
}

func getFuncId(pattern string, excludeId int, tests ...func(*T)) (id int) {
	id = -1

	for i, test := range tests {
		funcValue := reflect.NewValue(test)

		switch f := funcValue.(type) {
		case *reflect.FuncValue:
			funcName := runtime.FuncForPC(f.Get()).Name()
			matched, err := regexp.MatchString(pattern, funcName)
			if err == nil && matched && i != excludeId {
				id = i
			}
		}
	}

	return
}

func printCallerName() {
	pc, _, _, _ := runtime.Caller(2)
	callerName := runtime.FuncForPC(pc).Name()
	fmt.Printf("\n%s:\n", callerName)
}

func run(t *testing.T, format Formatter, tests ...func(*T)) {
	beforeAllFuncId := getFuncId(".*\\.beforeAll.*$", -1, tests...)
	afterAllFuncId := getFuncId(".*\\.afterAll.*$", -1, tests...)
	setupFuncId := getFuncId(".*\\.before.*$", beforeAllFuncId, tests...)
	teardownFuncId := getFuncId(".*\\.after.*$", afterAllFuncId, tests...)

	if beforeAllFuncId >= 0 {
		assertions := &T{t, STATUS_PASS, STATUS_PASS, false, &callerInfo{"", "", 0}}
		tests[beforeAllFuncId](assertions)
	}

	for i, test := range tests {

		assertions := &T{t, STATUS_PASS, STATUS_PASS, false, &callerInfo{"", "", 0}}

		if i == beforeAllFuncId {
			continue
		}

		if i == afterAllFuncId {
			continue
		}

		if i == setupFuncId || i == teardownFuncId {
			continue
		}

		if setupFuncId >= 0 {
			tests[setupFuncId](assertions)
		}

		test(assertions)

		if teardownFuncId >= 0 {
			tests[teardownFuncId](assertions)
		}

		if !assertions.Dry {
			format.PrintStatus(assertions.Status, assertions.callerInfo)
		} else {
			format.PrintDry(assertions.callerInfo)
		}
	}

	if afterAllFuncId >= 0 {
		assertions := &T{t, STATUS_PASS, STATUS_PASS, false, &callerInfo{"", "", 0}}
		tests[afterAllFuncId](assertions)
	}
}

// Run tests.
func Run(t *testing.T, tests ...func(*T)) {
	printCallerName()
	run(t, &TDDFormatter{}, tests...)
}

// Run tests but don't emit output and don't fail on failing
// assertions.
func DryRun(t *testing.T, tests ...func(*T)) {
	for _, test := range tests {
		test(&T{t, STATUS_PASS, STATUS_PASS, true, nil})
	}
}

// Run tests in BDD style
func Describe(t *testing.T, description string, tests ...func(*T)) {
	fmt.Printf("\n%s\n", description)
	run(t, &BDDFormatter{}, tests...)
}
