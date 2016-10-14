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

package prettytest

import (
	gocheck "gopkg.in/check.v1"
	"io/ioutil"
	"os"
	"testing"
)

var state, beforeState, afterState, beforeAllState, afterAllState int

type testSuite struct{ Suite }

type beforeAfterSuite struct{ Suite }
type bddFormatterSuite struct{ Suite }

func (suite *testSuite) TestNoAssertions() {}

func (suite *testSuite) TestFailMessage() {
	suite.True(false)
	suite.True(false, "This should fail with a custom error message")
	suite.MustFail()
}

func (suite *testSuite) TestTrue() {
	suite.True(true)
	suite.Not(suite.True(false))
}

func (suite *testSuite) TestError() {
	suite.Error("This test should be marked as failed")
	suite.MustFail()
}

func (suite *testSuite) TestNot() {
	suite.Not(suite.Equal("foo", "bar"))
	suite.Not(suite.True(false))
}

func (suite *testSuite) TestFalse() {
	suite.False(false)
	suite.Not(suite.False(true))
}

func (suite *testSuite) TestEqual() {
	suite.Equal("foo", "foo")
}

func (suite *testSuite) TestCheck() {
	suite.Check("42", gocheck.Equals, "42")
	suite.Check("42", gocheck.Equals, "43")
	suite.Check("notnil", gocheck.IsNil, "custom error")
	suite.MustFail()
}

func (suite *testSuite) TestNil() {
	var v *int = nil
	suite.Nil(v)
	suite.Nil(nil)
	suite.Not(suite.Nil([]byte{1, 2, 3}))
}

func (suite *testSuite) TestPath() {
	ioutil.WriteFile("./testfile", nil, 0600)
	suite.Path("testfile")
	suite.Not(suite.Path("foo"))
}

func (suite *testSuite) TestPending() {
	suite.Pending()
}

func (suite *testSuite) After() {
	os.Remove("testfile")
}

func (suite *beforeAfterSuite) Before() {
	state += 2
	beforeState++
}

func (suite *beforeAfterSuite) After() {
	state--
	afterState--
}

func (suite *beforeAfterSuite) BeforeAll() {
	state = 0
	beforeAllState++
}

func (suite *beforeAfterSuite) AfterAll() {
	state = 0
	afterAllState--
}

func (suite *beforeAfterSuite) TestSetup_1() {
	suite.Equal(2, state)
}

func (suite *beforeAfterSuite) TestSetup_2() {
	suite.Equal(3, state)
}

func TestPrettyTest(t *testing.T) {
	Run(
		t,
		new(testSuite),
		new(beforeAfterSuite),
	)
	if beforeAllState != 1 {
		t.Errorf("beforeAllState should be 1 after all tests but was %d\n", beforeAllState)
	}
	if afterAllState != -1 {
		t.Errorf("afterAllState should be -1 after all tests but was %d\n", afterAllState)
	}
}

func (suite *bddFormatterSuite) Should_use_green_on_passing_examples() {
	suite.True(true)
}

func (suite *bddFormatterSuite) Should_use_yellow_on_pending_examples() {
	suite.Pending()
}

func (suite *bddFormatterSuite) Should_use_yellow_on_examples_with_no_assertions() {}

func TestBDDStyleSpecs(t *testing.T) {
	RunWithFormatter(
		t,
		&BDDFormatter{Description: "BDD Formatter"},
		new(bddFormatterSuite),
	)
}
