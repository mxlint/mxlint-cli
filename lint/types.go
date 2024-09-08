package lint

import "encoding/xml"

type TestSuites struct {
	XMLName    xml.Name    `xml:"testsuites" json:"-"`
	Testsuites []Testsuite `xml:"testsuite" json:"testsuites"`
	Rules      []Rule      `xml:"-" json:"rules"`
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

type Rule struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Severity    string `json:"severity"`
	RuleNumber  string `json:"ruleNumber"`
	Remediation string `json:"remediation"`
	RuleName    string `json:"ruleName"`
	Path        string `json:"path"`
	SkipReason  string `json:"skipReason"`
	Pattern     string `json:"pattern"`
	PackageName string `json:"packageName"`
}
