package main

import "testing"

func TestLaunchGuard(t *testing.T) {
	closeFirst, err := lockLaunch()
	if err != nil {
		t.Fatal(err)
	}
	closed := false
	defer func() {
		if !closed {
			closeFirst()
		}
	}()
	if closeSecond, err := lockLaunch(); err == nil {
		closeSecond()
		t.Fatal("Concurrent launch accepted")
	}
	closeFirst()
	closed = true
	closeThird, err := lockLaunch()
	if err != nil {
		t.Fatal("Launch remains locked", err)
	}
	closeThird()
	if err = requireGameClosed(); err != nil {
		t.Fatal(err)
	}
}
