// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mbox_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	. "github.com/korylprince/mbox"
)

// slowReader is a reader that returns only a few bytes at a time, to test the incremental
// reads in Scanner.Scan.
type slowReader struct {
	max int
	buf io.Reader
}

func (sr *slowReader) Read(p []byte) (n int, err error) {
	if len(p) > sr.max {
		p = p[0:sr.max]
	}
	return sr.buf.Read(p)
}

var testError = errors.New("testError")

// Test the correct error is returned when the split function errors out.
func TestSplitError(t *testing.T) {
	// Create a split function that delivers a little data, then a predictable error.
	numSplits := 0
	const okCount = 7
	errorSplit := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF {
			panic("didn't get enough data")
		}
		if numSplits >= okCount {
			return 0, nil, testError
		}
		numSplits++
		return 1, data[0:1], nil
	}
	// Read the data.
	const text = "abcdefghijklmnopqrstuvwxyz"
	buf := strings.NewReader(text)
	s := NewScanner(&slowReader{1, buf})
	s.Split(errorSplit)
	var i int
	for i = 0; s.Scan(); i++ {
		if len(s.Bytes()) != 1 || text[i] != s.Bytes()[0] {
			t.Errorf("#%d: expected %q got %q", i, text[i], s.Bytes()[0])
		}
	}
	// Check correct termination location and error.
	if i != okCount {
		t.Errorf("unexpected termination; expected %d tokens got %d", okCount, i)
	}
	err := s.Err()
	if err != testError {
		t.Fatalf("expected %q got %v", testError, err)
	}
}

// Test for issue 5268.
type alwaysError struct{}

func (alwaysError) Read(p []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestNonEOFWithEmptyRead(t *testing.T) {
	scanner := NewScanner(alwaysError{})
	for scanner.Scan() {
		t.Fatal("read should fail")
	}
	err := scanner.Err()
	if err != io.ErrUnexpectedEOF {
		t.Errorf("unexpected error: %v", err)
	}
}

// Test that Scan finishes if we have endless empty reads.
type endlessZeros struct{}

func (endlessZeros) Read(p []byte) (int, error) {
	return 0, nil
}

func TestBadReader(t *testing.T) {
	scanner := NewScanner(endlessZeros{})
	for scanner.Scan() {
		t.Fatal("read should fail")
	}
	err := scanner.Err()
	if err != io.ErrNoProgress {
		t.Errorf("unexpected error: %v", err)
	}
}

// Test that empty tokens, including at end of line or end of file, are found by the scanner.
// Issue 8672: Could miss final empty token.

func commaSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	for i := 0; i < len(data); i++ {
		if data[i] == ',' {
			return i + 1, data[:i], nil
		}
	}
	if !atEOF {
		return 0, nil, nil
	}
	return 0, data, nil
}

func TestEmptyTokens(t *testing.T) {
	s := NewScanner(strings.NewReader("1,2,3,"))
	values := []string{"1", "2", "3", ""}
	s.Split(commaSplit)
	var i int
	for i = 0; i < len(values); i++ {
		if !s.Scan() {
			break
		}
		if s.Text() != values[i] {
			t.Errorf("%d: expected %q got %q", i, values[i], s.Text())
		}
	}
	if i != len(values) {
		t.Errorf("got %d fields, expected %d", i, len(values))
	}
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}
}

func loopAtEOFSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) > 0 {
		return 1, data[:1], nil
	}
	return 0, data, nil
}

func TestDontLoopForever(t *testing.T) {
	s := NewScanner(strings.NewReader("abc"))
	s.Split(loopAtEOFSplit)
	// Expect a panic
	defer func() {
		err := recover()
		if err == nil {
			t.Fatal("should have panicked")
		}
		if msg, ok := err.(string); !ok || !strings.Contains(msg, "empty tokens") {
			panic(err)
		}
	}()
	for count := 0; s.Scan(); count++ {
		if count > 1000 {
			t.Fatal("looping")
		}
	}
	if s.Err() != nil {
		t.Fatal("after scan:", s.Err())
	}
}

type countdown int

func (c *countdown) split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if *c > 0 {
		*c--
		return 1, data[:1], nil
	}
	return 0, nil, nil
}

// Check that the looping-at-EOF check doesn't trigger for merely empty tokens.
func TestEmptyLinesOK(t *testing.T) {
	c := countdown(10000)
	s := NewScanner(strings.NewReader(strings.Repeat("\n", 10000)))
	s.Split(c.split)
	for s.Scan() {
	}
	if s.Err() != nil {
		t.Fatal("after scan:", s.Err())
	}
	if c != 0 {
		t.Fatalf("stopped with %d left to process", c)
	}
}
