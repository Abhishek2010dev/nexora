package nexora

//
// func TestNewContext(t *testing.T) {
// 	nexora := &Nexora{
// 		trees: make(map[string]*node),
// 		pool:  &sync.Pool{},
// 	}
//
// 	ctx := NewContext(nexora)
// 	ctx.init(nil, nil, nil)
// 	if ctx.nexora != nexora {
// 		t.Errorf("Expected nexora to be set in context, got %v", ctx.nexora)
// 	}
// 	if ctx.index != -1 {
// 		t.Errorf("Expected index to be -1, got %d", ctx.index)
// 	}
// 	if ctx.handlers != nil {
// 		t.Errorf("Expected handlers to be nil, got %v", ctx.handlers)
// 	}
// }
//
// func TestContextNext(t *testing.T) {
// 	handlerCalled := false
// 	nexora := &Nexora{
// 		trees: make(map[string]*node),
// 		pool:  &sync.Pool{},
// 	}
//
// 	ctx := NewContext(nexora)
//
// 	ctx.init(nil, nil, Params{}, func(c *Context) error {
// 		handlerCalled = true
// 		return nil
// 	})
//
// 	if err := ctx.Next(); err != nil {
// 		t.Errorf("Expected no error, got %v", err)
// 	}
//
// 	if !handlerCalled {
// 		t.Error("Expected handler to be called, but it was not")
// 	}
// }
//
// func TestContextWithParams(t *testing.T) {
// 	params := Params{
// 		{"id", "123"},
// 		{"name", "test"},
// 	}
// 	ctx := NewContext(nil)
// 	ctx.init(nil, nil, params)
//
// 	if ctx.Param("id") != "123" {
// 		t.Errorf("Expected param 'id' to be '123', got %s", ctx.Param("id"))
// 	}
//
// 	if ctx.Param("name") != "test" {
// 		t.Errorf("Expected param 'name' to be 'test', got %s", ctx.Param("name"))
// 	}
//
// 	if ctx.Param("nonexistent") != "" {
// 		t.Error("Expected param 'nonexistent' to be empty")
// 	}
// }
//
// func TestContextAbort(t *testing.T) {
// 	handlerCalled := false
// 	ctx := NewContext(nil)
// 	ctx.init(nil, nil, Params{}, func(c *Context) error {
// 		handlerCalled = true
// 		c.Abort() // Abort the context
// 		return nil
// 	}, func(c *Context) error {
// 		t.Error("This handler should not be called")
// 		return nil
// 	})
//
// 	if err := ctx.Next(); err != nil {
// 		t.Errorf("Expected no error, got %v", err)
// 	}
//
// 	if !handlerCalled {
// 		t.Error("Expected first handler to be called, but it was not")
// 	}
// }
