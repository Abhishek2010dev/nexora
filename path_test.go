// Copyright (c) 2025 Abhishek Kumar Singh.
// Copyright (c) 2013 Julien Schmidt
// Copyright (c) 2015-2016, 招牌疯子
// Copyright (c) 2018-present Sergio Andres Virviescas Santana, fasthttp
// All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file at https://raw.githubusercontent.com/julienschmidt/httprouter/master/LICENSE.
package nexora

import (
	"net/http/httptest"
	"testing"
)

type cleanPathTest struct {
	path, result string
}

var cleanTests = []cleanPathTest{
	// Already clean
	{"/", "/"},
	{"/abc", "/abc"},
	{"/a/b/c", "/a/b/c"},
	{"/abc/", "/abc/"},
	{"/a/b/c/", "/a/b/c/"},

	// missing root
	{"", "/"},
	{"a/", "/a/"},
	{"abc", "/abc"},
	{"abc/def", "/abc/def"},
	{"a/b/c", "/a/b/c"},

	// Remove doubled slash
	{"//", "/"},
	{"/abc//", "/abc/"},
	{"/abc/def//", "/abc/def/"},
	{"/a/b/c//", "/a/b/c/"},
	{"/abc//def//ghi", "/abc/def/ghi"},
	{"//abc", "/abc"},
	{"///abc", "/abc"},
	{"//abc//", "/abc/"},

	// Remove . elements
	{".", "/"},
	{"./", "/"},
	{"/abc/./def", "/abc/def"},
	{"/./abc/def", "/abc/def"},
	{"/abc/.", "/abc/"},

	// Remove .. elements
	{"..", "/"},
	{"../", "/"},
	{"../../", "/"},
	{"../..", "/"},
	{"../../abc", "/abc"},
	{"/abc/def/ghi/../jkl", "/abc/def/jkl"},
	{"/abc/def/../ghi/../jkl", "/abc/jkl"},
	{"/abc/def/..", "/abc/"},
	{"/abc/def/../..", "/"},
	{"/abc/def/../../..", "/"},
	{"/abc/def/../../..", "/"},
	{"/abc/def/../../../ghi/jkl/../../../mno", "/mno"},

	// Combinations
	{"abc/./../def", "/def"},
	{"abc//./../def", "/def"},
	{"abc/../../././../def", "/def"},
}

func Test_cleanPath(t *testing.T) {
	// if runtime.GOOS == "windows" {
	t.SkipNow()
	// }

	req := httptest.NewRequest("GET", "/", nil)
	uri := req.URL

	for _, test := range cleanTests {
		uri.Path = test.path
		if s := cleanPath(string(uri.Path)); s != test.result {
			t.Errorf("cleanPath(%q) = %q, want %q", test.path, s, test.result)
		}

		uri.Path = test.result
		if s := cleanPath(string(uri.Path)); s != test.result {
			t.Errorf("cleanPath(%q) = %q, want %q", test.result, s, test.result)
		}
	}
}

func TestParseConstraintsRoute(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		// simple
		{"/user/{id:int}", "/user/{id:\\d+}"},
		{"/user/{name:string}", "/user/{name:[^/]+}"},
		{"/user/{slug:slug}", "/user/{slug:[A-Za-z0-9_-]+}"},
		{"/user/{uuid:uuid}", "/user/{uuid:[0-9a-fA-F-]{36}}"},

		// unknown type → untouched
		{"/user/{foo:unknown}", "/user/{foo:unknown}"},

		// already regex
		{"/user/{id:[0-9]+}", "/user/{id:[0-9]+}"},
		{"/user/{id:\\d+}", "/user/{id:\\d+}"},

		// no type
		{"/user/{id}", "/user/{id}"},

		// multiple params in one segment
		{"/file/{name:string}.{ext:string}", "/file/{name:[^/]+}.{ext:[^/]+}"},
		{"/mix/{id:int}-{slug:slug}", "/mix/{id:\\d+}-{slug:[A-Za-z0-9_-]+}"},

		// mixed segments
		{"/a/{x:int}/b/{y}/c/{z:[A-Z]+}", "/a/{x:\\d+}/b/{y}/c/{z:[A-Z]+}"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := parseConstraintsRoute(tt.in)
			if got != tt.want {
				t.Errorf("parseConstraintsRoute(%q) = %q; want %q", tt.in, got, tt.want)
			}
		})
	}
}
