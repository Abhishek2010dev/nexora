package nexora

import (
	"net/http/httptest"
	"testing"
)

func TestContext_BindParams(t *testing.T) {
	type User struct {
		Name   string `param:"name"`
		ID     int    `param:"id"`
		Admin  bool   `param:"admin"`
		Height float64 `param:"height"`
	}

	req := httptest.NewRequest("GET", "/users/123", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	ctx.params = map[string]string{
		"name":   "john",
		"id":     "123",
		"admin":  "true",
		"height": "1.83",
	}

	var user User
	if err := ctx.BindParams(&user); err != nil {
		t.Fatalf("BindParams failed: %v", err)
	}

	if user.Name != "john" {
		t.Errorf("Name = %q; want %q", user.Name, "john")
	}
	if user.ID != 123 {
		t.Errorf("ID = %d; want %d", user.ID, 123)
	}
	if user.Admin != true {
		t.Errorf("Admin = %v; want %v", user.Admin, true)
	}
    if user.Height != 1.83 {
		t.Errorf("Height = %f; want %f", user.Height, 1.83)
	}
}

func TestContext_BindParams_ErrorCases(t *testing.T) {
	ctx := newContext(nil)

	t.Run("non-pointer", func(t *testing.T) {
		type User struct{}
		var user User
		if err := ctx.BindParams(user); err == nil {
			t.Error("expected an error for non-pointer, but got nil")
		}
	})

	t.Run("non-struct pointer", func(t *testing.T) {
		var i int
		if err := ctx.BindParams(&i); err == nil {
			t.Error("expected an error for non-struct pointer, but got nil")
		}
	})

	t.Run("unsupported-type", func(t *testing.T) {
		type User struct {
			Complex complex128 `param:"complex"`
		}
		ctx.params = map[string]string{"complex": "1+2i"}
		var user User
		if err := ctx.BindParams(&user); err == nil {
			t.Error("expected an error for unsupported type, but got nil")
		}
	})
}
