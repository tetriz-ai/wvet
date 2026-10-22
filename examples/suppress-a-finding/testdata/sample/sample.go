// Package sample shows where a directive takes effect and where it does not.
package sample

import (
	"errors"
	"sync"
)

func lookup() (string, error) { return "", errors.New("miss") }

// OnTheCallLine is suppressed: swallowcheck reads the call's line.
func OnTheCallLine() string {
	v, err := lookup() //swallowcheck:ignore // a miss means the default
	if err != nil {
		return "default"
	}
	return v
}

// OnTheCheckLine is suppressed: swallowcheck also reads the nil check's line.
func OnTheCheckLine() string {
	v, err := lookup()
	if err != nil { //nolint:swallowcheck // a miss means the default
		return "default"
	}
	return v
}

// AboveTheStatement is suppressed: the directive is alone on the line above.
func AboveTheStatement(work func()) {
	// work recovers its own panics.
	//goroutinecheck:ignore
	go work()
}

// TrailingTheStatement is reported: goroutinecheck reads only the line above.
func TrailingTheStatement(work func()) {
	go work() //goroutinecheck:ignore
}

// AboveTheWait is suppressed by a bare nolint, which covers every analyzer.
func AboveTheWait(wg *sync.WaitGroup) {
	//nolint // every Done is deferred at the top of a short job
	wg.Wait()
}

// TwoLinesAbove is reported: a blank line separates the directive from the Wait.
func TwoLinesAbove(wg *sync.WaitGroup) {
	//waitgroupcheck:ignore

	wg.Wait()
}
