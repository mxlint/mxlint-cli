package lint

import "encoding/xml"

type TestSuites struct {
	XMLName    xml.Name `xml:"testsuites"`
	Testsuites []Testsuite
}

type Testsuite struct {
	XMLName   xml.Name `xml:"testsuite"`
	Name      string   `xml:"name,attr"`
	Tests     int      `xml:"tests,attr"`
	Failures  int      `xml:"failures,attr"`
	Skipped   int      `xml:"skipped,attr"`
	Time      float64  `xml:"time,attr"`
	Testcases []Testcase
}

type Testcase struct {
	XMLName xml.Name `xml:"testcase"`
	Name    string   `xml:"name,attr"`
	Time    float64  `xml:"time,attr"`
	Failure *Failure `xml:"failure,omitempty"`
	Skipped *Skipped `xml:"skipped,omitempty"`
}

type Failure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Data    string `xml:",chardata"`
}
type Skipped struct {
	Message string `xml:"message,attr"`
}
