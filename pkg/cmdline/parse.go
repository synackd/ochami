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
	return parseToStruct(string(raw))
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

func parseToStruct(input string) *CmdLine {
	var (
		last      *paramItem
		ll        *paramItem
		llTracker     = ll
		numParams int = 0
	)
	keyMap := make(map[string][]*paramItem)
	doParse(input, func(flag, key, canonicalKey, value, trimmedValue string) {
		newParam := Param{
			CanonicalKey: canonicalKey,
			Key:          key,
			Raw:          flag,
			Value:        trimmedValue,
		}
		newParamItem := &paramItem{
			param: newParam,
		}
		if llTracker == nil {
			// Linked list is empty, create first item
			ll = newParamItem
			llTracker = ll
		} else {
			// Linked list is nonempty, append item and set
			// prev/next pointers
			newParamItem.prev = llTracker
			llTracker.next = newParamItem
			llTracker = llTracker.next
			keyMap[canonicalKey] = append(keyMap[canonicalKey], llTracker)
		}
		numParams++
		keyMap[canonicalKey] = append(keyMap[canonicalKey], newParamItem)
		last = newParamItem
	})
	return &CmdLine{
		last:      last,
		list:      ll,
		keyMap:    keyMap,
		numParams: numParams,
	}
}
