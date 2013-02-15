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
	"io/ioutil"
	"os"
	"testing"
)

var state, beforeState, afterState, beforeAllState, afterAllState int

type testSuite struct{ Suite }
type beforeAfterSuite struct{ Suite }
type bddFormatterSuite struct{ Suite }

func (suite *testSuite) TestNoAssertions() {}

func (suite *testSuite) TestTrueFalse() {
	suite.True(true)
	suite.False(false)
	suite.True(suite.False(false))
}

func (suite *testSuite) TestEqualNotEqual() {
	suite.Equal("foo", "foo")
	suite.NotEqual("foo", "bar")
	suite.False(suite.Equal("foo", "bar"))
	suite.False(suite.NotEqual("foo", "foo"))
}

func (suite *testSuite) TestNil() {
	var v *int = nil
	suite.Nil(v)
	suite.Nil(nil)
	suite.True(suite.Nil(nil))
}

func (suite *testSuite) TestNotNil() {
	suite.NotNil([]byte{1, 2, 3})
	suite.True(suite.NotNil(1))
}

func (suite *testSuite) TestPath() {
	ioutil.WriteFile("testfile", nil, 0600)
	suite.Path("testfile")
	suite.False(suite.Path("foo"))
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
