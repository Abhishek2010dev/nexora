package nexora

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	
)

func TestContext_BindQuery(t *testing.T) {
	type Query struct {
		Name string `query:"name"`
		Age  int    `query:"age"`
	}

	n := New()
	req := httptest.NewRequest(http.MethodGet, "/?name=test&age=10", nil)
	c := n.pool.Get().(*Context)
	c.init(req, httptest.NewRecorder())

	var q Query
	if err := c.BindQuery(&q); err != nil {
		t.Errorf("BindQuery() error = %v", err)
	}

	if q.Name != "test" {
		t.Errorf("expected name to be 'test', got %s", q.Name)
	}

	if q.Age != 10 {
		t.Errorf("expected age to be 10, got %d", q.Age)
	}
}

func TestContext_BindForm(t *testing.T) {
	type Form struct {
		Name string `form:"name"`
		Age  int    `form:"age"`
	}

	n := New()
	form := "name=test&age=10"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c := n.pool.Get().(*Context)
	c.init(req, httptest.NewRecorder())

	var f Form
	if err := c.BindForm(&f); err != nil {
		t.Errorf("BindForm() error = %v", err)
	}

	if f.Name != "test" {
		t.Errorf("expected name to be 'test', got %s", f.Name)
	}

	if f.Age != 10 {
		t.Errorf("expected age to be 10, got %d", f.Age)
	}
}

func TestContext_BindQueryAndForm(t *testing.T) {
	type QueryAndForm struct {
		Name string `query:"name" form:"name"`
		Age  int    `query:"age" form:"age"`
	}

	n := New()

	// Test with query parameters
	req := httptest.NewRequest(http.MethodGet, "/?name=test_query&age=10", nil)
	c := n.pool.Get().(*Context)
	c.init(req, httptest.NewRecorder())

	var qf QueryAndForm
	if err := c.BindQuery(&qf); err != nil {
		t.Errorf("BindQuery() error = %v", err)
	}

	if qf.Name != "test_query" {
		t.Errorf("expected name to be 'test_query', got %s", qf.Name)
	}

	if qf.Age != 10 {
		t.Errorf("expected age to be 10, got %d", qf.Age)
	}

	// Test with form data
	form := "name=test_form&age=20"
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.init(req, httptest.NewRecorder())

	if err := c.BindForm(&qf); err != nil {
		t.Errorf("BindForm() error = %v", err)
	}

	if qf.Name != "test_form" {
		t.Errorf("expected name to be 'test_form', got %s", qf.Name)
	}

	if qf.Age != 20 {
		t.Errorf("expected age to be 20, got %d", qf.Age)
	}
}


