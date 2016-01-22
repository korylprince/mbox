package mbox_test

//go:generate go-bindata -pkg mbox_test -o data_test.go test_data/

import (
	"fmt"
	"testing"

	. "github.com/korylprince/mbox"
)

var separatorTests = []struct {
	text   string
	idx    int
	length int
}{
	{ //good
		"From 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxx Mon Apr 20 02:27:10 2015\n",
		0,
		96,
	},
	{ //good; military time
		"From 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxx Mon Apr 20 16:27:10 2015\n",
		0,
		96,
	},
	{ //good; multiple
		"From 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxx Mon Apr 20 02:27:10 2015\nstuff...\nFrom 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxx Mon Apr 20 02:27:10 2015\n",
		0,
		96,
	},
	{ //good; extra front data
		"From <garbage data...\nsup?\nblahFrom 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxx Mon Apr 20 02:27:10 2015\n",
		31,
		96,
	},
	{ //good; extra end data
		"From 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxx Mon Apr 20 02:27:10 2015\nsome more data...\nhey\n\n",
		0,
		96,
	},
	{ //good; suspect (close but not correct line) front data
		"From 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxx Mon Apr 20 From 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxx Mon Apr 20 02:27:10 2015\n",
		82,
		96,
	},
	{ //bad addr-spec
		"From 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox Mon Apr 20 02:27:10 2015\n",
		-1,
		0,
	},
	{ //bad timestamp
		"From 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxx Mon Apr 20 02:10 2015\n",
		-1,
		0,
	},
	{ //missing first space
		"From1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxx Mon Apr 20 02:27:10 2015\n",
		-1,
		0,
	},
	{ //missing second space
		"From 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxxMon Apr 20 02:27:10 2015\n",
		-1,
		0,
	},
	{ //missing second space and date
		"From 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox",
		-1,
		0,
	},
	{ //missing eol
		"From 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxxMon Apr 20 02:27:10 2015",
		-1,
		0,
	},
	{ //good crlf
		"From 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxx Mon Apr 20 02:27:10 2015\r\n",
		0,
		97,
	},
	{ //good crlf; multiple
		"From 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxx Mon Apr 20 02:27:10 2015\r\nstuff...\r\nFrom 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxx Mon Apr 20 02:27:10 2015\r\n",
		0,
		97,
	},
	{ //good crlf; extra front data
		"From <garbage data...\r\nsup?\r\nblahFrom 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxx Mon Apr 20 02:27:10 2015\r\n",
		33,
		97,
	},
	{ //good crlf; extra end data
		"From 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxx Mon Apr 20 02:27:10 2015\r\nsome more data...\r\nhey\r\n\r\n",
		0,
		97,
	},
	{ //good crlf; suspect (close but not correct line) front data
		"From 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxx Mon Apr 20 From 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxx Mon Apr 20 02:27:10 2015\r\n",
		82,
		97,
	},
	{ //bad addr-spec crlf
		"From 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox Mon Apr 20 02:27:10 2015\r\n",
		-1,
		0,
	},
	{ //bad timestamp crlf
		"From 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxx Mon Apr 20 02:10 2015\r\n",
		-1,
		0,
	},
	{ //missing first space crlf
		"From1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxx Mon Apr 20 02:27:10 2015\r\n",
		-1,
		0,
	},
	{ //missing second space crlf
		"From 1498981366547862741-8cd918dd-8bbb-49f5-a4df-032f222154c9.mbox@xxxMon Apr 20 02:27:10 2015\r\n",
		-1,
		0,
	},
}

func TestFindSeparator(t *testing.T) {
	for _, test := range separatorTests {
		i, l := FindSeparator([]byte(test.text))
		if i != test.idx {
			t.Errorf("Got idx: %v, Expected: %v", i, test.idx)
		}
		if l != test.length {
			t.Errorf("Got length: %v, Expected: %v", l, test.length)
		}
	}
}

var mboxTests = []struct {
	atEOF   bool
	advance int
	err     error
}{
	{ //good, two messages
		false,
		752,
		nil,
	},
	{ //eof
		true,
		0,
		ErrorUnexpectedEOF,
	},
	{ //good; eof and no data left
		true,
		0,
		nil,
	},
	{ //good, one message and eof
		true,
		752,
		nil,
	},
	{ //good, one message and no eof (request more data for next separator line)
		false,
		0,
		nil,
	},
	{ //good, need more data
		false,
		0,
		nil,
	},
}

func TestScanMessage(t *testing.T) {
	for i, test := range mboxTests {
		data := MustAsset(fmt.Sprintf("test_data/%d.mbox", i))
		res := MustAsset(fmt.Sprintf("test_data/%d.res", i))
		a, d, e := ScanMessage(data, test.atEOF)

		if a != test.advance {
			t.Errorf("Got advance: %v, Expected: %v", a, test.advance)
		}

		if string(d) != string(res) {
			t.Errorf("Got data: %s, Expected: %s", d, res)
		}

		if e != test.err {
			t.Errorf("Got err: %v, Expected: %v", e, test.err)
		}
	}
}
