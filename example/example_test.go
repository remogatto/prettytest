package example

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


// Your tests start here
func (t *testSuite) TestTrueIsTrue() {
	t.True(true)
}

func (t *testSuite) TestEquality() {
	t.Equal("awesome", "awesome")
}

func (t *testSuite) TestComposition() {
	t.False(t.Path("foo"))
}


//failing test
func (t *testSuite) TestInequality() {
	t.NotEqual("awesome", "awesome")
}
