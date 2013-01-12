# What's that?

PrettyTest is a simple testing library without bells & whistles that
produces pretty outputs.

![PrettyTest](http://remogatto.github.com/images/prettytest.png)

# Features

PrettyTest main features are:

  * A simple assertion vocabulary for better readability
  * Customizable formatters through interfaces
  * It's integrated with the go test command
  * It has pretty and colorful output with reports

# Quick start

<pre>
package foo

import (
	"github.com/remogatto/prettytest"
	"testing"
)

type testSuite struct {
	prettytest.Suite
}

func (t *testSuite) TestTrueIsTrue() {
	t.True(true)
}

func TestFoo(t *testing.T) {
	prettytest.RunWithFormatter(
		t,
		new(prettytest.TDDFormatter),
		new(testSuite),
	)
}
</pre>

Then, to run the tests simply use the the go test command:

<pre>
$ go test

testSuite:
	OK	TestTrueIsTrue                (1 assertion(s))

1 tests, 1 passed, 0 failed, 0 pending
</pre>

# LICENSE

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
