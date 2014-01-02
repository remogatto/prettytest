package prettytest

import (
	"fmt"
	"path/filepath"
	"strings"
)

const formatTag = "\t%s\t"

var (
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

type FinalReport struct {
	Passed, Failed, ExpectedFailures, Pending, NoAssertions int
}

func (r *FinalReport) Total() int {
	return r.Passed + r.Failed + r.ExpectedFailures + r.Pending + r.NoAssertions
}

// Formatter is the interface each formatter should implement.
type Formatter interface {
	PrintSuiteInfo(suite *Suite)
	PrintStatus(testFunc *TestFunc)
	PrintFinalReport(report *FinalReport)
	PrintErrorLog(errorLog []*Error)

	// AllowedMethodPattern returns a regexp for the allowed
	// method name (e.g. "^Test.*" for the TDDFormatter)
	AllowedMethodsPattern() string
}

/*TDDFormatter is a very simple TDD-like formatter.

Legend:

* F  - Test Failed

* OK - Test Passed

* EF - An Expected Failure occured

* NA - Not Assertions found

* PE - Pending test
*/
type TDDFormatter struct{}

func (formatter *TDDFormatter) PrintSuiteInfo(suite *Suite) {
	fmt.Printf("\n%s.%s:\n\n", suite.Package, suite.Name)
}

func (formatter *TDDFormatter) PrintStatus(testFunc *TestFunc) {
	callerName := testFunc.Name
	switch testFunc.Status {
	case STATUS_FAIL:
		fmt.Printf(formatTag+"%-30s(%d assertion(s))\n", labelFAIL, callerName, len(testFunc.Assertions))
	case STATUS_MUST_FAIL:
		fmt.Printf(formatTag+"%-30s(%d assertion(s))\n", labelMUSTFAIL, callerName, len(testFunc.Assertions))
	case STATUS_PASS:
		fmt.Printf(formatTag+"%-30s(%d assertion(s))\n", labelPASS, callerName, len(testFunc.Assertions))
	case STATUS_PENDING:
		fmt.Printf(formatTag+"%-30s(%d assertion(s))\n", labelPENDING, callerName, len(testFunc.Assertions))
	case STATUS_NO_ASSERTIONS:
		fmt.Printf(formatTag+"%-30s(%d assertion(s))\n", labelNOASSERTIONS, callerName, len(testFunc.Assertions))

	}
}

func (formatter *TDDFormatter) PrintErrorLog(logs []*Error) {
	if len(logs) > 0 {
		currentTestFuncHeader := ""
		for _, error := range logs {
			if currentTestFuncHeader != error.TestFunc.Name {
				fmt.Printf("\n%s:\n", error.TestFunc.Name)
			}
			filename := filepath.Base(error.Assertion.Filename)
			fmt.Printf("\t(%s:%d) %s\n", filename, error.Assertion.Line, error.Assertion.ErrorMessage)
			currentTestFuncHeader = error.TestFunc.Name
		}
	}
}

func (formatter *TDDFormatter) PrintFinalReport(report *FinalReport) {
	fmt.Printf("\n%d tests, %d passed, %d failed, %d expected failures, %d pending, %d with no assertions\n",
		report.Total(), report.Passed, report.Failed, report.ExpectedFailures, report.Pending, report.NoAssertions)
}

func (formatter *TDDFormatter) AllowedMethodsPattern() string {
	return "^Test.*"
}

// BDDFormatter is a formatter à la rspec.
type BDDFormatter struct {
	Description string
}

func (formatter *BDDFormatter) PrintSuiteInfo(suite *Suite) {
	fmt.Printf("\n%s:\n", formatter.Description)
}

func (formatter *BDDFormatter) PrintStatus(testFunc *TestFunc) {
	shouldText := strings.Replace(testFunc.Name, "_", " ", -1)
	switch testFunc.Status {
	case STATUS_FAIL:
		fmt.Printf("- %s\n", red(shouldText))
	case STATUS_PASS:
		fmt.Printf("- %s\n", green(shouldText))
	case STATUS_MUST_FAIL:
		fmt.Printf("- %s\n", green(shouldText))
	case STATUS_PENDING:
		fmt.Printf("- %s\t(Not Yet Implemented)\n", yellow(shouldText))
	case STATUS_NO_ASSERTIONS:
		fmt.Printf("- %s\t(No assertions found)\n", yellow(shouldText))
	}
}

func (formatter *BDDFormatter) PrintFinalReport(report *FinalReport) {
	fmt.Printf("\n%d examples, %d passed, %d failed, %d expected failures, %d pending, %d with no assertions\n",
		report.Total(),
		report.Passed,
		report.Failed,
		report.ExpectedFailures,
		report.Pending,
		report.NoAssertions)
}

func (formatter *BDDFormatter) PrintErrorLog(logs []*Error) {
	if len(logs) > 0 {
		currentTestFuncHeader := ""
		for _, error := range logs {
			if currentTestFuncHeader != error.TestFunc.Name {
				fmt.Printf("\n%s:\n", error.TestFunc.Name)
			}
			filename := filepath.Base(error.Assertion.Filename)
			fmt.Printf("\t(%s:%d) %s\n", filename, error.Assertion.Line, error.Assertion.ErrorMessage)
			currentTestFuncHeader = error.TestFunc.Name
		}
	}
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
