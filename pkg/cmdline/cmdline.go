// Use of this source code is governed by the LICENSE file in this module's root
// directory.

// Package cmdline is parser for kernel command-line arguments. It is based on
// github.com/u-root/u-root/pkg/cmdline, except that it can parse strings
// instead of having to read from a file.
//
// It's conformant with
// https://www.kernel.org/doc/html/v4.14/admin-guide/kernel-parameters.html,
// though making 'var_name' and 'var-name' equivalent may need to be done
// separately.
package cmdline

import (
	"fmt"
	"strings"
)

type paramItem struct {
	param Param
	next  *paramItem
	prev  *paramItem
}

type Param struct {
	CanonicalKey string
	Key          string
	Raw          string
	Value        string
}

func (p Param) String() string {
	return p.Raw
}

// CmdLine provides a way to easily parse through kernel command line arguments
type CmdLine struct {
	list      *paramItem              // Linked list of all parameters
	last      *paramItem              // Last param in linked list
	keyMap    map[string][]*paramItem // Map to linked list items for faster reference
	numParams int                     // Total parameter count
}

// NewCmdLine returns a pointer to a CmdLine struct parsed with line.
func NewCmdLine(line []byte) *CmdLine {
	return parse(line)
}

func (c *CmdLine) String() string {
	var s []string
	for llTracker := c.list; llTracker != nil; llTracker = llTracker.next {
		s = append(s, llTracker.param.String())
	}
	return strings.Join(s, " ")
}

// ContainsFlag verifies that the kernel cmdline has a flag set
func (c *CmdLine) ContainsFlag(flag string) bool {
	_, present := c.GetFlag(flag)
	return present
}

// GetFlag returns the values of a flag, and whether it was set
func (c *CmdLine) GetFlag(flag string) ([]string, bool) {
	canonicalFlag := strings.Replace(flag, "-", "_", -1)
	piPtrs, present := c.keyMap[canonicalFlag]
	var vals []string
	for _, p := range piPtrs {
		vals = append(vals, p.param.Value)
	}
	return vals, present
}

func (c *CmdLine) SetFlag(flag, value string) {
	canonicalFlag := strings.Replace(flag, "-", "_", -1)
	newParam := Param{
		Key:          flag,
		CanonicalKey: canonicalFlag,
		Value:        dequote(value),
	}
	if value == "" {
		newParam.Raw = flag
	} else {
		newParam.Raw = fmt.Sprintf("%s=%s", flag, value)
	}
	newParamItem := &paramItem{
		param: newParam,
	}
	if ptrList, exists := c.keyMap[canonicalFlag]; exists {
		first := true
		for _, ptr := range ptrList {
			if ptr == nil {
				continue
			}
			if first {
				newParamItem.prev = ptr.prev
				newParamItem.next = ptr.next
				if ptr.prev != nil {
					ptr.prev.next = newParamItem
				}
				if ptr.next != nil {
					ptr.next.prev = newParamItem
				} else {
					c.last = newParamItem
				}
				ptr.prev = nil
				ptr.next = nil
				first = false
			} else {
				if ptr.prev != nil {
					ptr.prev.next = ptr.next
				}
				if ptr.next != nil {
					ptr.next.prev = ptr.prev
				}
			}
		}
		c.keyMap[canonicalFlag] = []*paramItem{c.keyMap[canonicalFlag][0]}
	} else {
		c.last.next = newParamItem
		newParamItem.prev = c.last
		c.keyMap[canonicalFlag] = []*paramItem{newParamItem}
		c.last = newParamItem
	}
}

func (c *CmdLine) PrintLL() {
	for llTracker := c.list; llTracker != nil; llTracker = llTracker.next {
		fmt.Printf("%p: %s (prev=%p, next=%p)\n", llTracker, llTracker.param, llTracker.prev, llTracker.next)
	}
	fmt.Printf("last: %p\n", c.last)
}

// FlagsForModule gets all flags for a designated module and returns them as a
// space-seperated string designed to be passed to insmod. Note that similarly
// to flags, module names with - and _ are treated the same.
func (c *CmdLine) FlagsForModule(name string) string {
	var ret string
	flagsAdded := make(map[string]bool) // Ensures duplicate flags aren't both added
	// Module flags come as moduleName.flag in /proc/cmdline
	prefix := strings.Replace(name, "-", "_", -1) + "."
	for llTracker := c.list; llTracker != nil; llTracker = llTracker.next {
		canonicalFlag := strings.Replace(llTracker.param.Key, "-", "_", -1)
		if !flagsAdded[canonicalFlag] && strings.HasPrefix(canonicalFlag, prefix) {
			flagsAdded[canonicalFlag] = true
			// They are passed to insmod space seperated as flag=val
			ret += strings.TrimPrefix(canonicalFlag, prefix) + "=" + llTracker.param.Value + " "
		}
	}
	return ret
}
