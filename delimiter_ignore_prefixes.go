package stream2sentence

import "strings"

// Titles and abbreviations
var titlesAndAbbreviations = []string{
	"Mr.", "Mrs.", "Ms.", "Dr.", "Prof.", "Rev.", "St.",
	"Ph.D.", "Phd.", "PhD.", "M.D.", "B.A.", "M.A.", "D.D.S.", "J.D.",
	"Co.", "Corp.", "Ave.", "Blvd.", "Rd.", "Mt.",
	"a.m.", "p.m.", "Jr.", "Sr.",
	"Gov.", "Gen.", "Capt.", "Lt.", "Maj.", "Col.", "Adm.", "Cmdr.",
	"Sgt.", "Cpl.", "Pvt.", "U.S.", "U.K.", "vs.", "i.e.", "e.g.",
	"Vol.", "Art.", "Sec.", "Chap.", "Fig.", "Ref.", "Dept.",
}

// Dates and times
var datesAndTimes = []string{
	"Jan.", "Feb.", "Mar.", "Apr.", "Jun.", "Jul.", "Aug.",
	"Sep.", "Oct.", "Nov.", "Dec.",
	"Mon.", "Tue.", "Wed.", "Thu.", "Fri.", "Sat.", "Sun.",
}

// Financial abbreviations
var financialAbbreviations = []string{
	"Inc.", "Ltd.", "Corp.", "PLC.", "LLC.", "LLP.",
	"P/E.", "EPS.", "NAV.", "ROI.", "ROA.", "ROE.",
}

// Country abbreviations
var countryAbbreviations = []string{
	"U.S.A.", "U.K.", "U.A.E.", "P.R.C.", "D.R.C.", "R.O.C.",
	"E.U.", "U.N.", "A.U.",
	"U.S.", "U.K.", "E.U.", "P.R.C.", "D.R.C.", "R.O.C.",
}

// DelimiterIgnorePrefixes contains all prefixes that should ignore delimiters
var DelimiterIgnorePrefixes map[string]bool

// initDelimiterIgnorePrefixes initializes the DelimiterIgnorePrefixes map
func initDelimiterIgnorePrefixes() {
	if DelimiterIgnorePrefixes != nil {
		return
	}

	allPrefixes := []string{}
	allPrefixes = append(allPrefixes, titlesAndAbbreviations...)
	allPrefixes = append(allPrefixes, datesAndTimes...)
	allPrefixes = append(allPrefixes, financialAbbreviations...)
	allPrefixes = append(allPrefixes, countryAbbreviations...)

	DelimiterIgnorePrefixes = make(map[string]bool)
	for _, prefix := range allPrefixes {
		DelimiterIgnorePrefixes[prefix] = true
		DelimiterIgnorePrefixes[strings.ToLower(prefix)] = true
	}
}

// IsDelimiterIgnorePrefix checks if a prefix should ignore delimiters
func IsDelimiterIgnorePrefix(prefix string) bool {
	initDelimiterIgnorePrefixes()
	return DelimiterIgnorePrefixes[prefix] || DelimiterIgnorePrefixes[strings.ToLower(prefix)]
}
