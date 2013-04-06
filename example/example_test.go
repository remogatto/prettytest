package example

import (
	"testing"
	"github.com/remogatto/prettytest"
	"launchpad.net/gocheck"
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


// Your tests start here

func (t *testSuite) TestTrueIsTrue() {
	t.True(true)
}

func (t *testSuite) TestEquality() {
	t.Equal("awesome", "awesome")
}

func (t *testSuite) TestNot() {
	t.Not(t.Path("foo"))
}

func (t *testSuite) TestGoCheck() {
	t.Check("foo", gocheck.Equals, "foo")
}

// Failing tests

func (t *testSuite) TestMustFail() {
	t.Error("This test must fail.")
	t.MustFail()
}

func (t *testSuite) TestInequality() {
	t.Equal("awesome", "ugly")
	t.MustFail()
}
