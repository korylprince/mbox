[![Go Reference](https://pkg.go.dev/badge/github.com/korylprince/mbox.svg)](https://pkg.go.dev/github.com/korylprince/mbox)

This package provides a `ScanMessage` function that can be used with `bufio.Scanner`. This function splits data on RFC 4155 "default" separator lines, including the line in the data returned. The implementation could be more efficient, but it's pretty fast right now.

Since emails are often larger than `bufio.Scanner`'s token size limit of 64kB, a custom `mbox.Scanner` is provided, with a tunable `MaxTokenSize` (which defaults to 64MB.) This `Scanner` is just a stripped down version of the standard library version with a few changes (see below.)

If you have any issues or questions, email the email address below, or open an issue at:
https://github.com/korylprince/mbox/issues

# Changes from Standard Library `scan.go`

* Removed unneeded SplitFuncs
* Moved `maxConsecutiveEmptyReads` into `scan.go`
* Changed `MaxScanTokenSize` to 64MB from 64kB
* Made `Scanner.MaxTokenSize` public
* Changed the default `ScanFunc` for `NewScanner` to `ScanMessage`
* Changed the default `Scanner.buf` size to 1MB from 4kB (size of buffer at initialization)

# Usage

`godoc github.com/korylprince/mbox`

Or read the source. It's pretty simple and readable.

Example:

```go
f, err := os.Open("/path/to/mbox")
if err != nil {
	//do something with err
}
defer f.Close()

s := mbox.NewScanner(f)
s.MaxTokenSize = 1024 * 1024 * 1024 // 1GB max size, or whatever you want
for s.Scan() {
	b := s.Bytes()
	//copy bytes to buffer, otherwise they will be overwritten
	buf := make([]byte, len(b))
	copy(buf, b)
	r := bufio.NewReader(r)

	_, _, err = r.ReadLine() //read in mbox separator line
	if err != nil {
		//do something with err
	}

	msg, err := mail.ReadMessage(buf)
	if err != nil {
		//do something with err
	}

	//do something with msg!

if err := s.Err(); err != nil {
	//do something with err
}
```


# Copyright Information

`scan.go` and `scan_test.go` are modified files from the main Go distribution and thus retain the [Go Programming Language License](https://golang.org/LICENSE).

Test data is taken from http://mailman.postel.org/pipermail/touch-mm.mbox/touch-mm.mbox

All other code is Copyright 2021 Kory Prince (korylprince at gmail dot com) and licensed under the LICENSE provided with the code.
