// Copyright (c) 2025 Abhishek Kumar Singh.
// Copyright (c) 2013 Julien Schmidt
// Copyright (c) 2015-2016, 招牌疯子
// Copyright (c) 2018-present Sergio Andres Virviescas Santana, fasthttp
// All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file at https://raw.githubusercontent.com/julienschmidt/httprouter/master/LICENSE.
package nexora

import (
	"fmt"
	"regexp"
	"strings"

	gstrings "github.com/savsgio/gotils/strings"
)

// cleanPath removes the '.' if it is the last character of the route
func cleanPath(path string) string {
	return strings.TrimSuffix(path, ".")
}

// getOptionalPaths returns all possible paths when the original path
// has optional arguments
func getOptionalPaths(path string) []string {
	paths := make([]string, 0)

	start := 0
walk:
	for {
		if start >= len(path) {
			return paths
		}

		c := path[start]
		start++

		if c != '{' {
			continue
		}

		newPath := ""
		hasRegex := false
		questionMarkIndex := -1

		brackets := 0

		for end, c := range []byte(path[start:]) {
			switch c {
			case '{':
				brackets++

			case '}':
				if brackets > 0 {
					brackets--
					continue
				} else if questionMarkIndex == -1 {
					continue walk
				}

				end++
				newPath += path[questionMarkIndex+1 : start+end]

				path = path[:questionMarkIndex] + path[questionMarkIndex+1:] // remove '?'
				paths = append(paths, newPath)
				start += end - 1

				continue walk

			case ':':
				hasRegex = true

			case '?':
				if hasRegex {
					continue
				}

				questionMarkIndex = start + end
				newPath += path[:questionMarkIndex]

				if len(path[:start-2]) == 0 {
					// include the root slash because the param is in the first segment
					paths = append(paths, "/")
				} else if !gstrings.Include(paths, path[:start-2]) {
					// include the path without the wildcard
					// -2 due to remove the '/' and '{'
					paths = append(paths, path[:start-2])
				}
			}
		}
	}
}

// constraintsType maps supported parameter types to their corresponding regex patterns.
// These are strict patterns that match common Go primitive types.
var constraintsType = map[string]string{
	// Signed integers with strict ranges where practical
	"int":   `-?\d+`,
	"int8":  `-?(?:12[0-7]|1[01]\d|[1-9]?\d)`,                       // -128 to 127
	"int16": `-?(?:3276[0-7]|327[0-5]\d|32[0-6]\d{2}|[12]?\d{1,3})`, // -32768 to 32767
	"int32": `-?\d+`,                                                // full range too large for practical regex
	"int64": `-?\d+`,

	// Unsigned integers
	"uint":   `\d+`,
	"uint8":  `(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)`,                             // 0 to 255
	"uint16": `(?:6553[0-5]|655[0-2]\d|65[0-4]\d{2}|6[0-4]\d{3}|[1-5]?\d{1,4})`, // 0 to 65535
	"uint32": `\d+`,
	"uint64": `\d+`,

	// Floats
	"float32": `[-+]?\d*\.?\d+`,
	"float64": `[-+]?\d*\.?\d+`,

	// Common utility types
	"string": `[^/]+`,
	"slug":   `[A-Za-z0-9_-]+`,
	"uuid":   `[0-9a-fA-F-]{36}`,
	"alpha":  `[A-Za-z]+`,
	"alnum":  `[A-Za-z0-9]+`,
	"bool":   `(true|false|0|1)`,
	"path":   `.*`,

	// other
	"email":    `[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`,
	"ip":       `(?:\d{1,3}\.){3}\d{1,3}`,
	"ipv6":     `(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}`, // basic full IPv6 match
	"hostname": `[a-zA-Z0-9.-]+`,
	"date":     `\d{4}-\d{2}-\d{2}`,
	"time":     `\d{2}:\d{2}(?::\d{2})?`,
	"hex":      `[0-9a-fA-F]+`,
	"base64":   `[A-Za-z0-9+/=]+`,
}

// paramRegex matches path parameters of the form:
// {name:type}, {name:type(extra)}, {name}, etc.
var paramRegex = regexp.MustCompile(`{(\w+)(?::([a-zA-Z0-9_]+)(?:\(([0-9, ]+)\))?)?}`)

// parseConstraintsRoute parses a route pattern and converts
// typed parameters into their corresponding regex patterns.
//
// Supported forms:
//
//	/user/{id:int}
//	/range/{age:range(18,120)}
//	/len/{code:len(6)}
//	/min/{val:min(10)}
//	/max/{val:max(100)}
//
// Unknown types are left untouched for further handling.
//
// NOTE: `range()`, `min()`, and `max()` apply ONLY to integer types.
func parseConstraintsRoute(route string) string {
	segments := strings.Split(route, "/")
	out := make([]string, 0, len(segments))

	for _, seg := range segments {
		if seg == "" {
			continue
		}

		// Replace each {param:type(...)} occurrence
		newSeg := paramRegex.ReplaceAllStringFunc(seg, func(s string) string {
			m := paramRegex.FindStringSubmatch(s)
			if m == nil {
				return s
			}
			key, typ, extra := m[1], m[2], m[3]

			// Handle built-in constraints
			switch typ {
			case "max":
				if extra != "" {
					maxVal := parseInt(extra)
					if maxVal < 0 {
						panicf("nexora: invalid max() value in route: %s", s)
					}
					return fmt.Sprintf("{%s:%s}", key, generateMaxRegex(maxVal))
				}
			case "min":
				if extra != "" {
					minVal := parseInt(extra)
					if minVal < 0 {
						panicf("nexora: invalid min() value in route: %s", s)
					}
					return fmt.Sprintf("{%s:%s}", key, generateMinRegex(minVal))
				}
			case "len":
				if extra != "" {
					l := parseInt(extra)
					if l <= 0 || l > 1024 {
						panicf("nexora: invalid len() value in route: %s", s)
					}
					rgx := fmt.Sprintf(`[^/]{%d}`, l)
					return fmt.Sprintf("{%s:%s}", key, rgx)
				}
			case "range":
				if extra != "" {
					parts := strings.Split(extra, ",")
					if len(parts) == 2 {
						min := parseInt(strings.TrimSpace(parts[0]))
						max := parseInt(strings.TrimSpace(parts[1]))
						if min < 0 || max < 0 || max < min {
							panicf("nexora: invalid range() values in route: %s", s)
						}
						return fmt.Sprintf("{%s:%s}", key, generateRangeRegex(min, max))
					} else {
						panicf("nexora: invalid range() format in route: %s", s)
					}
				}
			default:
				if typ != "" {
					if regex, ok := constraintsType[typ]; ok {
						return fmt.Sprintf("{%s:%s}", key, regex)
					}
				}
			}
			// No type provided: leave as-is
			return s
		})

		out = append(out, newSeg)
	}

	return "/" + strings.Join(out, "/")
}

// generateRangeRegex builds a regex that matches any integer between min and max inclusive.
// If the range is too large, it falls back to a generic \d+ for performance.
func generateRangeRegex(min, max int) string {
	if max < min {
		panicf("nexora: invalid range, max (%d) is less than min (%d)", max, min)
	}
	if max-min > 1000 {
		return `\d+`
	}
	patterns := make([]string, 0, max-min+1)
	for i := min; i <= max; i++ {
		patterns = append(patterns, fmt.Sprintf("%d", i))
	}
	return "^(" + strings.Join(patterns, "|") + ")$"
}

// generateMinRegex builds a regex that matches any integer >= min, up to min+500.
// Beyond that, it falls back to a generic pattern to avoid huge regex.
func generateMinRegex(min int) string {
	if min < 0 {
		min = 0
	}
	if min > 100000 {
		return `\d+`
	}
	patterns := make([]string, 0, 501)
	for i := min; i <= min+500; i++ {
		patterns = append(patterns, fmt.Sprintf("%d", i))
	}
	return "^(" + strings.Join(patterns, "|") + ")$"
}

// generateMaxRegex builds a regex that matches any integer from 0 up to max.
// Beyond 1000, it falls back to a generic pattern.
func generateMaxRegex(max int) string {
	if max < 0 {
		max = 0
	}
	if max > 1000 {
		return `\d+`
	}
	patterns := make([]string, 0, max+1)
	for i := 0; i <= max; i++ {
		patterns = append(patterns, fmt.Sprintf("%d", i))
	}
	return "^(" + strings.Join(patterns, "|") + ")$"
}

// parseInt parses a string into an int, trimming whitespace.
// Returns 0 if parsing fails.
func parseInt(s string) int {
	var v int
	fmt.Sscanf(strings.TrimSpace(s), "%d", &v)
	return v
}
