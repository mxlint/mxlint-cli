package lint

import "encoding/xml"

type TestSuites struct {
	XMLName    xml.Name    `xml:"testsuites" json:"-"`
	Testsuites []Testsuite `xml:"testsuite" json:"testsuites"`
}

type Testsuite struct {
	XMLName   xml.Name   `xml:"testsuite" json:"-"`
	Name      string     `xml:"name,attr" json:"name"`
	Tests     int        `xml:"tests,attr" json:"tests"`
	Failures  int        `xml:"failures,attr" json:"failures"`
	Skipped   int        `xml:"skipped,attr" json:"skipped"`
	Time      float64    `xml:"time,attr" json:"time"`
	Testcases []Testcase `xml:"testcase" json:"testcases"`
}

type Testcase struct {
	XMLName xml.Name `xml:"testcase" json:"-"`
	Name    string   `xml:"name,attr" json:"name"`
	Time    float64  `xml:"time,attr" json:"time"`
	Failure *Failure `xml:"failure,omitempty" json:"failure,omitempty"`
	Skipped *Skipped `xml:"skipped,omitempty" json:"skipped,omitempty"`
}

type Failure struct {
	Message string `xml:"message,attr" json:"message"`
	Type    string `xml:"type,attr" json:"type"`
	Data    string `xml:",chardata" json:"-"`
}

type Skipped struct {
	Message string `xml:"message,attr" json:"message"`
}
