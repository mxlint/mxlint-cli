package lint

import "encoding/xml"

type PolicyResult int

const (
	PolicyResultPass PolicyResult = iota
	PolicyResultFail
	PolicyResultUnknown
)

type Testsuite struct {
	XMLName   xml.Name `xml:"testsuite"`
	Name      string   `xml:"name,attr"`
	Tests     int      `xml:"tests,attr"`
	Failures  int      `xml:"failures,attr"`
	Time      float64  `xml:"time,attr"`
	Testcases []Testcase
}

type Testcase struct {
	XMLName xml.Name `xml:"testcase"`
	Name    string   `xml:"name,attr"`
	Time    float64  `xml:"time,attr"`
	Failure *Failure `xml:"failure,omitempty"`
}

type Failure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Data    string `xml:",chardata"`
}
