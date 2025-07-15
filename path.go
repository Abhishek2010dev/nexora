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

var constraintsType = map[string]string{
	"int":      `\d+`,                                              // 123
	"string":   `[^/]+`,                                            // anything except "/"
	"slug":     `[A-Za-z0-9_-]+`,                                   // URL-friendly
	"uuid":     `[0-9a-fA-F-]{36}`,                                 // UUID v4 style
	"alpha":    `[A-Za-z]+`,                                        // letters only
	"alnum":    `[A-Za-z0-9]+`,                                     // letters and digits
	"float":    `\d+\.\d+`,                                         // simple floating-point
	"hex":      `[0-9a-fA-F]+`,                                     // hex digits
	"year":     `\d{4}`,                                            // 4-digit year
	"month":    `(0[1-9]|1[0-2])`,                                  // 01–12
	"day":      `(0[1-9]|[12][0-9]|3[01])`,                         // 01–31
	"bool":     `(true|false|0|1)`,                                 // booleans
	"username": `[A-Za-z0-9_]{3,16}`,                               // 3–16 chars
	"email":    `[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`, // basic email
	"phone":    `\+?[0-9]{7,15}`,                                   // simple international phone
}

var paramRegex = regexp.MustCompile(`{(\w+)(?::([^}]+))?\}`)

func parseConstraintsRoute(route string) string {
	sagments := strings.Split(route, "/")
	out := make([]string, 0, len(sagments))

	for _, seg := range sagments {
		if seg == "" {
			continue
		}

		newSeg := paramRegex.ReplaceAllStringFunc(seg, func(s string) string {
			matchs := paramRegex.FindStringSubmatch(s)
			key := matchs[1]
			typ := matchs[2]

			// If no type, leave as-is
			if typ == "" {
				return fmt.Sprintf("{%s}", key)
			}

			// If it's already a regex (e.g. starts with \ or [), keep it
			if strings.ContainsAny(typ, `\[]()^$`) {
				return fmt.Sprintf("{%s:%s}", key, typ)
			}

			regex, ok := constraintsType[typ]
			if !ok {
				return s // fallback: don't touch
			}
			return fmt.Sprintf("{%s:%s}", key, regex)
		})

		out = append(out, newSeg)
	}
	return "/" + strings.Join(out, "/")
}
