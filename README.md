# What's that?

PrettyTest is a simple testing library without bells & whistles that
produces pretty outputs.

![PrettyTest](http://remogatto.github.com/images/prettytest.png)

# Features

PrettyTest main features are:

  * A simple assertion vocabulary for better readability
  * Customizable formatters through interfaces
  * It's integrated with the go test command
  * It can use [gocheck](http://labix.org/gocheck) checkers (experimental)
  * It has pretty and colorful output with reports

# Quick start

~~~go
package foo

import (
	"github.com/remogatto/prettytest"
	"testing"
)

// Start of setup
type testSuite struct {
	prettytest.Suite
}

func TestRunner(t *testing.T) {
	prettytest.RunWithFormatter(
		t,
		new(prettytest.TDDFormatter),
		new(testSuite),
	)
}
// End of setup


//your tests start here
func (t *testSuite) TestTrueIsTrue() {
	t.True(true)
}

func (t *testSuite) TestEquality() {
	t.Equal("awesome", "awesome")
}

func (t *testSuite) TestNegation() {
	t.Not(t.Equal("awesome", "not good"))
}


//failing test

func (t *testSuite) TestInequality() {
	t.Equal("awesome", "pretty")
}
~~~

Then, to run the tests simply use the the go test command:

~~~bash
$ go test

testSuite:
	OK	TestComposition               (2 assertion(s))
	OK	TestEquality                  (1 assertion(s))
	F	TestInequality                (1 assertion(s))
	OK	TestTrueIsTrue                (1 assertion(s))

4 tests, 3 passed, 1 failed, 0 pending, 0 with no assertions
--- FAIL: TestRunner (0.00 seconds)
prettytest.go:453: 	Expected awesome to be not equal to awesome -- /home/andrea/src/sandbox/go/prettytest/example_test.go:39
		
FAIL
exit status 1
FAIL	_/home/andrea/src/sandbox/go/prettytest	0.014s
~~~

# TDD formatter legend

* F  - Test Failed
* OK - Test Passed
* EF - An Expected Failure occured
* NA - Not Assertions found
* PE - Pending test

# PrettyAutoTest

PrettyAutoTest is a command that continously watches for changes in
your source directory and - in case - re-executes the
tests. PrettyAutoTest promotes an agile development approach: write
tests, write the implementation that satisfies the tests, have an
immediate visual feedback about your progress.

Check this video: http://youtu.be/B35N6q3sveQ

In order to install PrettyAutoTest simply type:

~~~bash
$ go get -v github.com/remogatto/prettytest/pta/
~~~

Then simply launch the <tt>pta</tt> command move in your working
directory and type:

~~~bash
$ pta
~~~

Type <tt>pta -h</tt> for additional help.

# LICENSE

Copyright (c) 2013 Andrea Fazzi

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
