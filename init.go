// +build !android

package prettytest

import (
	"flag"
	"regexp"
)

var (
	testToRun = flag.String("pt.run", "", "[prettytest] regular expression that filters tests and examples to run")
)

func filterMethod(name string) bool {
	ok, _ := regexp.MatchString(*testToRun, name)
	return ok
}

func init() {
	flag.Parse()
}
