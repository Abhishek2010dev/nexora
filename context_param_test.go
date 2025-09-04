package nexora

import (
	"net/http/httptest"
	"testing"
)

func TestContext_Param_Generic(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	ctx.params = map[string]string{
		"string":  "hello",
		"int":     "-42",
		"uint":    "42",
		"float":   "3.14",
		"bool":    "true",
		"invalid": "not-a-number",
	}

	t.Run("string", func(t *testing.T) {
		val, err := ParamAs[string](ctx, "string")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != "hello" {
			t.Errorf("got %q, want %q", val, "hello")
		}
	})

	t.Run("int", func(t *testing.T) {
		val, err := ParamAs[int](ctx, "int")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != -42 {
			t.Errorf("got %d, want %d", val, -42)
		}
	})

	t.Run("uint", func(t *testing.T) {
		val, err := ParamAs[uint](ctx, "uint")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != 42 {
			t.Errorf("got %d, want %d", val, 42)
		}
	})

	t.Run("float64", func(t *testing.T) {
		val, err := ParamAs[float64](ctx, "float")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != 3.14 {
			t.Errorf("got %f, want %f", val, 3.14)
		}
	})

	t.Run("bool", func(t *testing.T) {
		val, err := ParamAs[bool](ctx, "bool")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != true {
			t.Errorf("got %v, want %v", val, true)
		}
	})

	t.Run("invalid-int", func(t *testing.T) {
		_, err := ParamAs[int](ctx, "invalid")
		if err == nil {
			t.Error("expected an error, but got nil")
		}
	})

	t.Run("missing", func(t *testing.T) {
		_, err := ParamAs[int](ctx, "missing")
		if err == nil {
			t.Error("expected an error for missing key, but got nil")
		}
	})
}
