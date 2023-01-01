package mbox_test

import (
	"bufio"
	"bytes"
	"fmt"
	"net/mail"
	"os"

	"github.com/korylprince/mbox"
)

func ExampleScanner() {
	f, err := os.Open("/path/to/mbox")
	if err != nil {
		// do something with err
	}
	defer f.Close()

	s := mbox.NewScanner(f)
	s.MaxTokenSize = 1024 * 1024 * 1024 // 1GB max size, or whatever you want
	for s.Scan() {
		b := s.Bytes()
		// copy bytes to buffer, otherwise they will be overwritten
		buf := make([]byte, len(b))
		copy(buf, b)
		r := bufio.NewReader(bytes.NewReader(b))

		_, _, err = r.ReadLine() // read in mbox separator line
		if err != nil {
			// do something with err
		}

		msg, err := mail.ReadMessage(r)
		if err != nil {
			// do something with err
		}

		// do something with msg
		fmt.Println(msg.Header)

		if err := s.Err(); err != nil {
			// do something with err
		}
	}
}
