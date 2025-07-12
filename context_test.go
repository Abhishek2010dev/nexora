package nexora

import "testing"

func TestContext(t *testing.T) {
	ctx := NewContext(nil, nil, Params{}, func(c *Context) error {
		return nil
	})

	if ctx.Request() != nil {
		t.Errorf("Expected nil request, got %v", ctx.Request())
	}

	if ctx.ResponseWriter() != nil {
		t.Errorf("Expected nil response writer, got %v", ctx.ResponseWriter())
	}

	if len(ctx.Params()) != 0 {
		t.Errorf("Expected empty params, got %v", ctx.Params())
	}

	if err := ctx.Next(); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestContextWithHandlers(t *testing.T) {
	handlerCalled := false
	ctx := NewContext(nil, nil, Params{}, func(c *Context) error {
		handlerCalled = true
		return nil
	})

	if err := ctx.Next(); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !handlerCalled {
		t.Error("Expected handler to be called, but it was not")
	}
}

func TestContextWithParams(t *testing.T) {
	params := Params{
		{"id", "123"},
		{"name", "test"},
	}
	ctx := NewContext(nil, nil, params)

	if ctx.Param("id") != "123" {
		t.Errorf("Expected param 'id' to be '123', got %s", ctx.Param("id"))
	}

	if ctx.Param("name") != "test" {
		t.Errorf("Expected param 'name' to be 'test', got %s", ctx.Param("name"))
	}

	if ctx.Param("nonexistent") != "" {
		t.Error("Expected param 'nonexistent' to be empty")
	}
}

func TestContextAbort(t *testing.T) {
	handlerCalled := false
	ctx := NewContext(nil, nil, Params{}, func(c *Context) error {
		handlerCalled = true
		c.Abort() // Abort the context
		return nil
	}, func(c *Context) error {
		t.Error("This handler should not be called")
		return nil
	})

	if err := ctx.Next(); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !handlerCalled {
		t.Error("Expected first handler to be called, but it was not")
	}
}
