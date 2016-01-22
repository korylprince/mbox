package mbox

import (
	"bytes"
	"fmt"
	"net/mail"
	"time"
)

const ctimeFormat = "Mon Jan 02 15:04:05 2006"

//ErrorUnexpectedEOF signals that EOF was found before an expected separator line
var ErrorUnexpectedEOF = fmt.Errorf("Expected separator line, Got EOF")

var separatorPrefix = []byte("From ")

//FindSeparator returns the start index and length of the first RFC 4155 "default" compliant separator line:
//From <RFC 2822 "addr-spec"> <timestamp in UNIX ctime format><EOL marker>
//idx is negative a separator line is not found
func FindSeparator(data []byte) (idx, size int) {
	//find "From "
	var start int // F in "From "
	if start = bytes.Index(data, separatorPrefix); start < 0 {
		return -1, 0
	}

	addrStart := start + len(separatorPrefix) //first character after "From "

	//find end of addr-spec
	var addrEnd int
	if addrEnd = bytes.IndexByte(data[addrStart:], ' '); addrEnd < 0 {
		return -1, 0
	}

	//normalize addrEnd to data
	addrEnd = addrStart + addrEnd //space after addr-spec

	timeStart := addrEnd + 1 //first character of timestamp

	//verify addr-spec is correct
	if _, err := mail.ParseAddress(string(data[addrStart:addrEnd])); err != nil {
		i, size := FindSeparator(data[timeStart:])
		if i == -1 {
			return -1, 0
		}
		return timeStart + i, size
	}

	eol := []byte{'\n'}

	//find end of line marker
	var timeEnd int
	if timeEnd = bytes.Index(data[timeStart:], eol); timeEnd < 0 {
		return -1, 0
	}

	//detect a CRLF
	if data[timeStart+timeEnd-1] == '\r' {
		eol = []byte{'\r', '\n'}
		timeEnd--
	}

	//normalize timeEnd to data
	timeEnd = timeStart + timeEnd // first character of eol

	end := timeEnd + len(eol) // first character after eol

	//verify timestamp is correct
	if _, err := time.Parse(ctimeFormat, string(data[timeStart:timeEnd])); err != nil {
		i, size := FindSeparator(data[timeStart:])
		if i == -1 {
			return -1, 0
		}
		return timeStart + i, size
	}

	return start, end - start
}

//ScanMessage is a bufio.SplitFunc that splits the input on RFC 4155 "default" compliant separator lines
//returning the line with its message
func ScanMessage(data []byte, atEOF bool) (advance int, token []byte, err error) {
	var idx, size int
	if atEOF {
		if len(data) == 0 {
			return 0, nil, nil
		}
		idx, size = FindSeparator(data)
		if idx == -1 {
			return 0, nil, ErrorUnexpectedEOF
		}
	} else {
		idx, size = FindSeparator(data)
		if idx == -1 {
			return 0, nil, nil
		}
	}

	newidx, _ := FindSeparator(data[idx+size:])
	if newidx != -1 {
		//normalize to data
		newidx = idx + size + newidx
		return newidx, data[idx:newidx], nil
	}

	//rest of data is message
	if atEOF {
		return len(data), data[idx:], nil
	}

	return 0, nil, nil
}
