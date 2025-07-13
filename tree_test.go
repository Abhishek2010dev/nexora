package nexora

import (
	"testing"

	"github.com/valyala/bytebufferpool"
)

type testHandler struct {
	name string
}

func h(_ string) Handler {
	return func(ctx *Context) error {
		return nil
	}
}

func TestTree_StaticRoute(t *testing.T) {
	tree := &tree{root: newNode("/")}
	tree.Add("/home", []Handler{h("homeHandler")})

	handlers, params, tsr := tree.Get("/home")
	if tsr {
		t.Errorf("unexpected TSR for /home")
	}
	if handlers == nil {
		t.Fatalf("expected handler, got nil")
	}
	if params != nil {
		t.Errorf("expected no params, got: %+v", params)
	}
}

func TestTree_ParamRoute(t *testing.T) {
	tree := &tree{root: newNode("/")}
	tree.Add("/user/{id}", []Handler{h("userHandler")})

	handlers, params, tsr := tree.Get("/user/123")
	if tsr {
		t.Errorf("unexpected TSR for /user/123")
	}
	if handlers == nil {
		t.Fatalf("expected handler, got nil")
	}
	if params["id"] != "123" {
		t.Errorf("expected param 'id' to be '123', got '%s'", params["id"])
	}
}

func TestTree_TrailingSlashRedirect(t *testing.T) {
	tree := &tree{root: newNode("/")}
	tree.Add("/about", []Handler{h("aboutHandler")})

	_, _, tsr := tree.Get("/about/")
	if !tsr {
		t.Errorf("expected TSR for /about/")
	}
}

func TestTree_CaseInsensitivePath(t *testing.T) {
	tree := &tree{root: newNode("/")}
	tree.Add("/Contact", []Handler{h("contactHandler")})

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	found := tree.FindCaseInsensitivePath("/contact", true, buf)
	if !found {
		t.Errorf("expected case-insensitive match for /contact")
	}
	if got := buf.String(); got != "/Contact" {
		t.Errorf("expected corrected path '/Contact', got '%s'", got)
	}
}

func TestTree_RegexRoute(t *testing.T) {
	tree := &tree{root: newNode("/")}

	// Add route with regex pattern
	tree.Add("/product/{pid:[0-9]+}", []Handler{h("productHandler")})

	// Match valid path
	handlers, params, tsr := tree.Get("/product/456")
	if tsr {
		t.Errorf("unexpected TSR for /product/456")
	}
	if handlers == nil {
		t.Fatalf("expected handler, got nil")
	}
	if params["pid"] != "456" {
		t.Errorf("expected param 'pid' to be '456', got '%s'", params["pid"])
	}

	// Match invalid path (non-numeric)
	handlers, params, tsr = tree.Get("/product/abc")
	if handlers != nil {
		t.Errorf("expected no handler for /product/abc")
	}
}

func TestTree_StaticOverParam(t *testing.T) {
	tree := &tree{root: newNode("/")}

	tree.Add("/conflict/{name}", []Handler{h("param")})
	tree.Add("/conflict/static", []Handler{h("static")})

	handlers, params, _ := tree.Get("/conflict/static")
	if handlers == nil {
		t.Fatalf("expected handler")
	}
	if len(params) != 0 {
		t.Errorf("expected no params for static match, got %+v", params)
	}
}

func TestTree_InlineSuffixParam(t *testing.T) {
	tree := &tree{root: newNode("/")}
	tree.Add("/{name}_world", []Handler{h("inlineParamHandler")})

	tests := []struct {
		path        string
		expected    bool
		expectedVal string
	}{
		{"/john_world", true, "john"},         // ✅ valid
		{"/alice_world", true, "alice"},       // ✅ valid
		{"/john_universe", false, ""},         // ❌ wrong suffix
		{"/johnworld", false, ""},             // ❌ missing separator
		{"/", false, ""},                      // ❌ root
		{"/john_worlds", false, ""},           // ❌ similar suffix
		{"/john-doe_world", true, "john-doe"}, // 🧪 special char (if allowed)
	}

	for _, tt := range tests {
		handlers, params, tsr := tree.Get(tt.path)
		if tsr {
			t.Errorf("unexpected TSR for %s", tt.path)
		}
		if tt.expected {
			if handlers == nil {
				t.Errorf("expected handler for %s", tt.path)
				continue
			}
			if params["name"] != tt.expectedVal {
				t.Errorf("expected param 'name' to be '%s' for %s, got '%s'", tt.expectedVal, tt.path, params["name"])
			}
		} else {
			if handlers != nil {
				t.Errorf("expected no handler for %s", tt.path)
			}
		}
	}
}

func TestTree_WildcardRoute(t *testing.T) {
	tree := &tree{root: newNode("/")}
	tree.Add("/static/{filepath:*}", []Handler{h("wildcardHandler")})

	handlers, params, _ := tree.Get("/static/css/style.css")
	if handlers == nil {
		t.Error("expected handler for wildcard path")
	}
	if params["filepath"] != "css/style.css" {
		t.Errorf("expected 'css/style.css', got '%s'", params["filepath"])
	}
}

func TestTree_OverlappingPrefix(t *testing.T) {
	tree := &tree{root: newNode("/")}
	tree.Add("/user", []Handler{h("userIndex")})
	tree.Add("/user/{id}", []Handler{h("userShow")})

	handlers, params, _ := tree.Get("/user")
	if handlers == nil {
		t.Fatal("expected handler for /user")
	}
	if len(params) != 0 {
		t.Errorf("expected no params for /user")
	}

	handlers, params, _ = tree.Get("/user/42")
	if handlers == nil {
		t.Fatal("expected handler for /user/42")
	}
	if params["id"] != "42" {
		t.Errorf("expected id=42, got %s", params["id"])
	}
}

func TestTree_MultiParamRoute(t *testing.T) {
	tree := &tree{root: newNode("/")}
	tree.Add("/{name}/{age}", []Handler{h("multiParamHandler")})

	tests := []struct {
		path      string
		wantName  string
		wantAge   string
		wantMatch bool
	}{
		{"/john/30", "john", "30", true},
		{"/alice/25", "alice", "25", true},
		{"/john/", "", "", false},         // missing age
		{"/john", "", "", false},          // incomplete
		{"/", "", "", false},              // empty
		{"/john/30/extra", "", "", false}, // too many segments
	}

	for _, tt := range tests {
		handlers, params, tsr := tree.Get(tt.path)
		if tsr {
			t.Errorf("unexpected TSR for path %s", tt.path)
		}
		if tt.wantMatch {
			if handlers == nil {
				t.Errorf("expected handler for path %s", tt.path)
				continue
			}
			if params["name"] != tt.wantName {
				t.Errorf("expected name=%s, got %s", tt.wantName, params["name"])
			}
			if params["age"] != tt.wantAge {
				t.Errorf("expected age=%s, got %s", tt.wantAge, params["age"])
			}
		} else {
			if handlers != nil {
				t.Errorf("expected no handler for path %s", tt.path)
			}
		}
	}
}

func TestTree_ConflictPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic due to handler conflict")
		}
	}()
	tree := &tree{root: newNode("/")}
	tree.Add("/conflict/{id}", []Handler{h("one")})
	tree.Add("/conflict/{id}", []Handler{h("two")}) // should panic
}

func TestTree_CaseInsensitiveTSR(t *testing.T) {
	tree := &tree{root: newNode("/")}
	tree.Add("/About/", []Handler{h("about")})

	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	found := tree.FindCaseInsensitivePath("/about", true, buf)
	if !found || buf.String() != "/About/" {
		t.Errorf("expected /About/ with TSR, got %s", buf.String())
	}
}

func BenchmarkTree_StaticRoute(b *testing.B) {
	tree := &tree{root: newNode("/")}
	tree.Add("/home", []Handler{h("homeHandler")})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Get("/home")
	}
}

func BenchmarkTree_ParamRoute(b *testing.B) {
	tree := &tree{root: newNode("/")}
	tree.Add("/user/{id}", []Handler{h("userHandler")})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Get("/user/123")
	}
}

func BenchmarkTree_RegexParamRoute(b *testing.B) {
	tree := &tree{root: newNode("/")}
	tree.Add("/order/{oid:[0-9]+}", []Handler{h("orderHandler")})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Get("/order/456")
	}
}

func BenchmarkTree_MultiParamRoute(b *testing.B) {
	tree := &tree{root: newNode("/")}
	tree.Add("/blog/{year}/{slug}", []Handler{h("blogHandler")})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Get("/blog/2024/zeno-rocks")
	}
}

func BenchmarkTree_InlineSuffixParam(b *testing.B) {
	tree := &tree{root: newNode("/")}
	tree.Add("/{name}_world", []Handler{h("inlineHandler")})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Get("/john_world")
	}
}

func BenchmarkTree_WildcardRoute(b *testing.B) {
	tree := &tree{root: newNode("/")}
	tree.Add("/assets/{filepath:*}", []Handler{h("wildHandler")})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Get("/assets/css/main.css")
	}
}
