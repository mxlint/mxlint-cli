package lint

import (
	"encoding/xml"
	"os"
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

func main() {
	ts := Testsuite{
		Name:     "ExampleSuite",
		Tests:    2,
		Failures: 1,
		Time:     0.123,
		Testcases: []Testcase{
			{Name: "Test1", Time: 0.123},
			{Name: "Test2", Time: 0.0, Failure: &Failure{Message: "Example failure", Type: "AssertionError", Data: "Expected true, got false"}},
		},
	}

	file, err := os.Create("xunit_report.xml")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	if err := encoder.Encode(ts); err != nil {
		panic(err)
	}
}
