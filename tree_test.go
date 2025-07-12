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

func handlerName(_ Handler) string {
	return "handler" // stub for testing, since we can't introspect functions
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
