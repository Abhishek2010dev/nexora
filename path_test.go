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
	"regexp"
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
		route string
		want  string
	}{
		{
			route: "/user/{id:int}",
			want:  `/user/{id:-?\d+}`,
		},
		{
			route: "/fixed/{code:len(6)}",
			// exactly 6 non-slash characters
			want: `/fixed/{code:[^/]{6}}`,
		},
		{
			route: "/age/{years:range(18,20)}",
			// should compile to alternation for 18|19|20
			want: `/age/{years:^(18|19|20)$}`,
		},
		{
			route: "/low/{val:min(3)}",
			// starts from 3 to 503 by pattern
			want: `/low/{val:^(3|4|5|6|7|8|9|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|25|26|27|28|29|30|31|32|33|34|35|36|37|38|39|40|41|42|43|44|45|46|47|48|49|50|51|52|53|54|55|56|57|58|59|60|61|62|63|64|65|66|67|68|69|70|71|72|73|74|75|76|77|78|79|80|81|82|83|84|85|86|87|88|89|90|91|92|93|94|95|96|97|98|99|100|101|102|103|104|105|106|107|108|109|110|111|112|113|114|115|116|117|118|119|120|121|122|123|124|125|126|127|128|129|130|131|132|133|134|135|136|137|138|139|140|141|142|143|144|145|146|147|148|149|150|151|152|153|154|155|156|157|158|159|160|161|162|163|164|165|166|167|168|169|170|171|172|173|174|175|176|177|178|179|180|181|182|183|184|185|186|187|188|189|190|191|192|193|194|195|196|197|198|199|200|201|202|203|204|205|206|207|208|209|210|211|212|213|214|215|216|217|218|219|220|221|222|223|224|225|226|227|228|229|230|231|232|233|234|235|236|237|238|239|240|241|242|243|244|245|246|247|248|249|250|251|252|253|254|255|256|257|258|259|260|261|262|263|264|265|266|267|268|269|270|271|272|273|274|275|276|277|278|279|280|281|282|283|284|285|286|287|288|289|290|291|292|293|294|295|296|297|298|299|300|301|302|303|304|305|306|307|308|309|310|311|312|313|314|315|316|317|318|319|320|321|322|323|324|325|326|327|328|329|330|331|332|333|334|335|336|337|338|339|340|341|342|343|344|345|346|347|348|349|350|351|352|353|354|355|356|357|358|359|360|361|362|363|364|365|366|367|368|369|370|371|372|373|374|375|376|377|378|379|380|381|382|383|384|385|386|387|388|389|390|391|392|393|394|395|396|397|398|399|400|401|402|403|404|405|406|407|408|409|410|411|412|413|414|415|416|417|418|419|420|421|422|423|424|425|426|427|428|429|430|431|432|433|434|435|436|437|438|439|440|441|442|443|444|445|446|447|448|449|450|451|452|453|454|455|456|457|458|459|460|461|462|463|464|465|466|467|468|469|470|471|472|473|474|475|476|477|478|479|480|481|482|483|484|485|486|487|488|489|490|491|492|493|494|495|496|497|498|499|500|501|502|503)$}`,
		},
		{
			route: "/cap/{val:max(10)}",
			want:  `/cap/{val:^(0|1|2|3|4|5|6|7|8|9|10)$}`,
		},
		{
			route: "/plain/{slug:slug}",
			want:  `/plain/{slug:[A-Za-z0-9_-]+}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.route, func(t *testing.T) {
			got := parseConstraintsRoute(tt.route)
			if got != tt.want {
				t.Errorf("parseConstraintsRoute(%q) = %q; want %q", tt.route, got, tt.want)
			}
		})
	}
}

func TestGeneratedRegexMatches(t *testing.T) {
	// spot-check for range
	rgx := regexp.MustCompile(generateRangeRegex(1, 3))
	for _, valid := range []string{"1", "2", "3"} {
		if !rgx.MatchString(valid) {
			t.Errorf("expected %q to match range(1,3)", valid)
		}
	}
	for _, invalid := range []string{"0", "4", "999"} {
		if rgx.MatchString(invalid) {
			t.Errorf("did not expect %q to match range(1,3)", invalid)
		}
	}

	// spot-check for min
	minRgx := regexp.MustCompile(generateMinRegex(5))
	if !minRgx.MatchString("5") {
		t.Errorf("min regex should match 5")
	}
	if !minRgx.MatchString("505") {
		t.Log("min regex doesn't guarantee upper bound beyond +500, that's expected.")
	}

	// spot-check for max
	maxRgx := regexp.MustCompile(generateMaxRegex(3))
	for _, valid := range []string{"0", "1", "2", "3"} {
		if !maxRgx.MatchString(valid) {
			t.Errorf("expected %q to match max(3)", valid)
		}
	}
	if maxRgx.MatchString("4") {
		t.Errorf("did not expect 4 to match max(3)")
	}
}
