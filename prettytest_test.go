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
	"testing"
	"os"
	"io/ioutil"
)

var state, beforeState, afterState, beforeAllState, afterAllState int

type testSuite struct { Suite }
type beforeAfterSuite struct { Suite }
type bddFormatterSuite struct { Suite }

func (suite *testSuite) testTrueFalse() {
	suite.True(true)
	suite.False(false)
}

func (suite *testSuite) testEqualNotEqual() {
	suite.Equal("foo", "foo")
	suite.NotEqual("foo", "bar")
}

func (suite *testSuite) testNil() { 
	suite.Nil(nil)
}

func (suite *testSuite) testNotNil() {
	suite.NotNil([]byte{1,2,3})
}

func (suite *testSuite) testPath() {
	ioutil.WriteFile("testfile", nil, 0600)
	suite.Path("testfile")
	// suite.Path("foo")
	// suite.True(suite.Failed())
}

func (suite *testSuite) testPending() { 
	suite.Pending()
}

func (suite *testSuite) after() {
	os.Remove("testfile")
}

func (suite *beforeAfterSuite) before() {
	state += 2
	beforeState++
}

func (suite *beforeAfterSuite) after() {
	state--
	afterState--
}

func (suite *beforeAfterSuite) beforeAll() {
	state = 0
	beforeAllState++
}

func (suite *beforeAfterSuite) afterAll() {
	state = 0
	afterAllState--
}

func (suite *beforeAfterSuite) testSetup_1() {
	suite.Equal(2, state)
}

func (suite *beforeAfterSuite) testSetup_2() {
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

func (suite *bddFormatterSuite) should_use_green_on_passing_examples() {
	suite.True(true)
}

func (suite *bddFormatterSuite) should_use_yellow_on_pending_examples() {
	suite.Pending()
}

func TestBDDStyleSpecs(t *testing.T) {
	RunWithFormatter(
		t,
		&BDDFormatter{Description: "BDD Formatter"},
		new(bddFormatterSuite),
	)
}
