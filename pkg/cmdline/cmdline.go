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
	"strings"
)

// CmdLine provides a way to easily parse through kernel command line arguments
type CmdLine struct {
	raw   string
	asMap map[string]string
}

// NewCmdLine returns a pointer to a CmdLine struct parsed with line.
func NewCmdLine(line []byte) *CmdLine {
	return parse(line)
}

// GetRaw returns the full, raw cmdline string
func (c *CmdLine) GetRaw() string {
	return c.raw
}

// GetMap returns the full map of the cmdline arguments
func (c *CmdLine) GetMap() map[string]string {
	return c.asMap
}

// ContainsFlag verifies that the kernel cmdline has a flag set
func (c *CmdLine) ContainsFlag(flag string) bool {
	_, present := c.GetFlag(flag)
	return present
}

// GetFlag returns the value of a flag, and whether it was set
func (c *CmdLine) GetFlag(flag string) (string, bool) {
	canonicalFlag := strings.Replace(flag, "-", "_", -1)
	value, present := c.asMap[canonicalFlag]
	return value, present
}

// FlagsForModule gets all flags for a designated module and returns them as a
// space-seperated string designed to be passed to insmod. Note that similarly
// to flags, module names with - and _ are treated the same.
func (c *CmdLine) FlagsForModule(name string) string {
	var ret string
	flagsAdded := make(map[string]bool) // Ensures duplicate flags aren't both added
	// Module flags come as moduleName.flag in /proc/cmdline
	prefix := strings.Replace(name, "-", "_", -1) + "."
	for flag, val := range c.asMap {
		canonicalFlag := strings.Replace(flag, "-", "_", -1)
		if !flagsAdded[canonicalFlag] && strings.HasPrefix(canonicalFlag, prefix) {
			flagsAdded[canonicalFlag] = true
			// They are passed to insmod space seperated as flag=val
			ret += strings.TrimPrefix(canonicalFlag, prefix) + "=" + val + " "
		}
	}
	return ret
}
