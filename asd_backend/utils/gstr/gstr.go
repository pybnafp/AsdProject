package gstr

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

func Split(str, delimiter string) []string {
	return strings.Split(str, delimiter)
}

func Replace(origin, search, replace string, count ...int) string {
	n := -1
	if len(count) > 0 {
		n = count[0]
	}
	return strings.Replace(origin, search, replace, n)
}

func Equal(a, b string) bool {
	return strings.EqualFold(a, b)
}

func Contains(str, substr string) bool {
	return strings.Contains(str, substr)
}

func SubStr(str string, start int, length ...int) (substr string) {
	lth := len(str)

	// Simple border checks.
	if start < 0 {
		start = 0
	}
	if start >= lth {
		start = lth
	}
	end := lth
	if len(length) > 0 {
		end = start + length[0]
		if end < start {
			end = lth
		}
	}
	if end > lth {
		end = lth
	}
	return str[start:end]
}

func Join(array []string, sep string) string {
	return strings.Join(array, sep)
}

func UcWords(str string) string {
	return strings.Title(str)
}

func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// ConvertBytesToReadableSize 将字节数转换为人类可读的单位
func ConvertBytesToReadableSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}

	sizes := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	base := math.Log(float64(bytes)) / math.Log(1024)
	size := math.Pow(1024, base-math.Floor(base))

	return fmt.Sprintf("%.2f %s", size, sizes[int(math.Floor(base))])
}

// FixPubMedLinks processes a string to find and fix PubMed link formats.
// It ensures that PubMed references are in the format: [PubMed: XXXXXXXX](https://pubmed.ncbi.nlm.nih.gov/XXXXXXXX)
func FixPubMedLinks(responseText string) string {
	// Phase 1: Handle bracketed PubMed references.
	// This regex finds patterns like "[PubMed: ID]" or "[PubMed: ID](any_url)".
	// Group 1: The full label part, e.g., "[PubMed: 12345]"
	// Group 2: The ID itself, e.g., "12345"
	// Group 3: (Optional) The existing URL part including parentheses, e.g., "(http://...)" or empty.
	//          The "(?:\s*(\([^\)]*\)))?" part means:
	//          an optional non-capturing group that contains:
	//              optional whitespace (\s*)
	//              followed by a capturing group (\([^\)]*\)) for the URL in parentheses.
	bracketedPubMedRegex, err := regexp.Compile(`(\[PubMed:\s*(\d+)\])(?:\s*(\([^\)]*\)))?`)
	if err != nil {
		panic(err)
	}

	responseText = bracketedPubMedRegex.ReplaceAllStringFunc(responseText, func(match string) string {
		subs := bracketedPubMedRegex.FindStringSubmatch(match)
		// subs[0] is the full matched string e.g. "[PubMed: 123](someurl)" or "[PubMed: 123]"
		// subs[1] is the label part e.g. "[PubMed: 123]" (this is what we want to keep for the label)
		// subs[2] is the ID e.g. "123"
		// subs[3] is the existing captured URL part including parentheses e.g. "(someurl)", or "" if not present

		pubmedID := subs[2]
		correctURL := fmt.Sprintf("(https://pubmed.ncbi.nlm.nih.gov/%s)", pubmedID)

		// If the captured URL part (subs[3]) exists and is already correct, return the original full match.
		if subs[3] != "" && subs[3] == correctURL {
			return match
		}

		// Otherwise, the URL is wrong, malformed, or missing.
		// Construct the correct full link using the captured label part (subs[1]) and the correct URL.
		return subs[1] + correctURL
	})

	// Phase 2: Handle plain "PubMed: ID" that are not already part of a correctly formatted markdown link.
	// To do this safely, we first protect already correct links with placeholders.

	var placeholders []string
	placeholderPrefix := "@@StellarCarePubMedFixedLink@@" // Unique placeholder prefix
	placeholderIndex := 0

	// Regex for already correctly formatted links (after Phase 1 or originally correct)
	correctLinkRegex, err := regexp.Compile(`\[PubMed:\s*\d+\]\(https://pubmed\.ncbi\.nlm\.nih\.gov/\d+\)`)
	if err != nil {
		panic(err)
	}

	// Substitute correct links with placeholders
	responseText = correctLinkRegex.ReplaceAllStringFunc(responseText, func(correctMatch string) string {
		placeholder := fmt.Sprintf("%s%d@@", placeholderPrefix, placeholderIndex)
		placeholders = append(placeholders, correctMatch)
		placeholderIndex++
		return placeholder
	})

	// Now, find and format any remaining plain "PubMed: ID" instances.
	// These are instances not caught by Phase 1 and not part of an already correct link.
	plainPubMedRegex, err := regexp.Compile(`PubMed:\s*(\d+)`)
	if err != nil {
		panic(err)
	}
	responseText = plainPubMedRegex.ReplaceAllStringFunc(responseText, func(match string) string {
		subs := plainPubMedRegex.FindStringSubmatch(match) // e.g., match is "PubMed: 12345"
		pubmedID := subs[1]                                // e.g., "12345"
		return fmt.Sprintf("[PubMed: %s](https://pubmed.ncbi.nlm.nih.gov/%s)", pubmedID, pubmedID)
	})

	// Restore the original correct links from placeholders
	// Iterate in reverse order of placeholder creation if needed, but simple replace should work with unique placeholders.
	for i, originalLink := range placeholders {
		placeholderToReplace := fmt.Sprintf("%s%d@@", placeholderPrefix, i)
		responseText = strings.Replace(responseText, placeholderToReplace, originalLink, 1)
	}

	return responseText
}
