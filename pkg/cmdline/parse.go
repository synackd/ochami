// Use of this source code is governed by the LICENSE file in this module's root
// directory.

package cmdline

import (
	"strings"
	"unicode"
)

// parse parses the raw byte slice into a CmdLine struct and returns a pointer
// to it.
func parse(raw []byte) *CmdLine {
	line := &CmdLine{
		// These work because string([]byte{}) is ""
		raw:   strings.TrimRight(string(raw), "\n"),
		asMap: parseToMap(string(raw)),
	}
	return line
}

func dequote(line string) string {
	if len(line) == 0 {
		return line
	}

	quotationMarks := `"'`

	var quote byte
	if strings.ContainsAny(string(line[0]), quotationMarks) {
		quote = line[0]
		line = line[1 : len(line)-1]
	}

	var context []byte
	var newLine []byte
	for _, c := range []byte(line) {
		if c == '\\' {
			context = append(context, c)
		} else if c == quote {
			if len(context) > 0 {
				last := context[len(context)-1]
				if last == c {
					context = context[:len(context)-1]
				} else if last == '\\' {
					// Delete one level of backslash
					newLine = newLine[:len(newLine)-1]
					context = []byte{}
				}
			} else {
				context = append(context, c)
			}
		} else if len(context) > 0 && context[len(context)-1] == '\\' {
			// If backslash is being used to escape something other
			// than "the quote", ignore it.
			context = []byte{}
		}

		newLine = append(newLine, c)
	}
	return string(newLine)
}

func doParse(input string, handler func(flag, key, canonicalKey, value, trimmedValue string)) {
	lastQuote := rune(0)
	quotedFieldsCheck := func(c rune) bool {
		switch {
		case c == lastQuote:
			lastQuote = rune(0)
			return false
		case lastQuote != rune(0):
			return false
		case unicode.In(c, unicode.Quotation_Mark):
			lastQuote = c
			return false
		default:
			return unicode.IsSpace(c)
		}
	}

	for _, flag := range strings.FieldsFunc(string(input), quotedFieldsCheck) {
		// kernel variables must allow '-' and '_' to be equivalent in variable
		// names. We will replace dashes with underscores for processing.

		// Split the flag into a key and value, setting value="1" if none
		split := strings.Index(flag, "=")

		if len(flag) == 0 {
			continue
		}
		var key, value string
		if split == -1 {
			key = flag
			value = "1"
		} else {
			key = flag[:split]
			value = flag[split+1:]
		}
		canonicalKey := strings.Replace(key, "-", "_", -1)
		trimmedValue := dequote(value)

		// Call the user handler
		handler(flag, key, canonicalKey, value, trimmedValue)
	}
}

// parseToMap turns a space-separated kernel commandline into a map
func parseToMap(input string) map[string]string {
	flagMap := make(map[string]string)
	doParse(input, func(flag, key, canonicalKey, value, trimmedValue string) {
		// We store the value twice, once with dash, once with underscores
		// Just in case people check with the wrong method
		flagMap[canonicalKey] = trimmedValue
		flagMap[key] = trimmedValue
	})

	return flagMap
}
