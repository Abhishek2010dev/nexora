package nexora

import (
	"encoding/json"
	"encoding/xml"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContext_SendString(t *testing.T) {
	// Create a test HTTP request and response recorder
	req := httptest.NewRequest("GET", "/hello", nil)
	rec := httptest.NewRecorder()

	// Create and initialize a Context
	ctx := newContext(nil)
	ctx.init(req, rec)

	// Call SendString
	err := ctx.SendString("Hello, Nexora!")
	if err != nil {
		t.Fatalf("SendString failed: %v", err)
	}

	// Check response
	result := rec.Body.String()
	if result != "Hello, Nexora!" {
		t.Errorf("unexpected response body: got %q, want %q", result, "Hello, Nexora!")
	}
}

func TestContext_Status(t *testing.T) {
	req := httptest.NewRequest("GET", "/status", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	ctx.Status(http.StatusTeapot).SendString("I'm a teapot")

	if rec.Code != http.StatusTeapot {
		t.Errorf("unexpected status code: got %d, want %d", rec.Code, http.StatusTeapot)
	}

	if body := rec.Body.String(); body != "I'm a teapot" {
		t.Errorf("unexpected body: got %q, want %q", body, "I'm a teapot")
	}
}

func TestContext_Next(t *testing.T) {
	req := httptest.NewRequest("GET", "/next", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	var called []string
	ctx.handlers = []Handler{
		func(c *Context) error {
			called = append(called, "1")
			return c.Next()
		},
		func(c *Context) error {
			called = append(called, "2")
			return nil
		},
	}

	err := ctx.Next()
	if err != nil {
		t.Fatalf("Next failed: %v", err)
	}

	want := []string{"1", "2"}
	if strings.Join(called, ",") != strings.Join(want, ",") {
		t.Errorf("handlers called in wrong order: got %v, want %v", called, want)
	}
}

func TestContext_Param(t *testing.T) {
	req := httptest.NewRequest("GET", "/users/42", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	// Simulate route parameters
	ctx.params = map[string]string{
		"id":   "42",
		"name": "",
	}

	// Test existing param
	id := ctx.Param("id")
	if id != "42" {
		t.Errorf("Param id = %q; want %q", id, "42")
	}

	// Test missing param with default
	role := ctx.Param("role", "admin")
	if role != "admin" {
		t.Errorf("Param role with default = %q; want %q", role, "admin")
	}

	// Test missing param without default
	role = ctx.Param("role")
	if role != "" {
		t.Errorf("Param role without default = %q; want empty string", role)
	}

	// Test empty param (should not use default)
	name := ctx.Param("name", "guest")
	if name != "" {
		t.Errorf("Param name = %q; want empty string", name)
	}
}

func TestContext_ParamExists(t *testing.T) {
	req := httptest.NewRequest("GET", "/items/5", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	ctx.params = map[string]string{
		"item": "5",
	}

	val, ok := ctx.ParamExists("item")
	if !ok || val != "5" {
		t.Errorf("ParamExists(item) = (%q, %v); want (%q, true)", val, ok, "5")
	}

	val, ok = ctx.ParamExists("missing")
	if ok || val != "" {
		t.Errorf("ParamExists(missing) = (%q, %v); want (\"\", false)", val, ok)
	}
}

func TestContext_Queries(t *testing.T) {
	req := httptest.NewRequest("GET", "/search?q=golang&tag=web&tag=fast&empty=", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	values := ctx.Queries()
	if values.Get("q") != "golang" {
		t.Errorf("Queries()[q] = %q; want %q", values.Get("q"), "golang")
	}
	if got := values["tag"]; len(got) != 2 || got[0] != "web" || got[1] != "fast" {
		t.Errorf("Queries()[tag] = %v; want [web fast]", got)
	}
	if _, ok := values["empty"]; !ok {
		t.Errorf("Queries()[empty] missing; want present")
	}
}

func TestContext_QueryArray(t *testing.T) {
	req := httptest.NewRequest("GET", "/search?tag=web&tag=fast", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	arr := ctx.QueryArray("tag")
	if len(arr) != 2 || arr[0] != "web" || arr[1] != "fast" {
		t.Errorf("QueryArray(tag) = %v; want [web fast]", arr)
	}

	arr = ctx.QueryArray("missing")
	if arr != nil && len(arr) != 0 {
		t.Errorf("QueryArray(missing) = %v; want nil or []", arr)
	}
}

func TestContext_Query(t *testing.T) {
	req := httptest.NewRequest("GET", "/search?q=golang", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	// existing key
	val := ctx.Query("q")
	if val != "golang" {
		t.Errorf("Query(q) = %q; want %q", val, "golang")
	}

	// missing key with default
	val = ctx.Query("page", "1")
	if val != "1" {
		t.Errorf("Query(page,1) = %q; want %q", val, "1")
	}

	// missing key without default
	val = ctx.Query("missing")
	if val != "" {
		t.Errorf("Query(missing) = %q; want \"\"", val)
	}
}

func TestContext_QueryExists(t *testing.T) {
	req := httptest.NewRequest("GET", "/search?q=golang&empty=", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	// key exists with value
	val, ok := ctx.QueryExists("q")
	if !ok || val != "golang" {
		t.Errorf("QueryExists(q) = (%q, %v); want (%q, true)", val, ok, "golang")
	}

	// key exists with empty value
	val, ok = ctx.QueryExists("empty")
	if !ok || val != "" {
		t.Errorf("QueryExists(empty) = (%q, %v); want (\"\", true)", val, ok)
	}

	// key does not exist
	val, ok = ctx.QueryExists("missing")
	if ok || val != "" {
		t.Errorf("QueryExists(missing) = (%q, %v); want (\"\", false)", val, ok)
	}
}

func TestContext_Port_HTTP(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com:8080/foo", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	if got := ctx.Port(); got != "8080" {
		t.Errorf("Port() = %q, want %q", got, "8080")
	}
}

func TestContext_Port_Defaults(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	// No port in Host should default to 80 (non-TLS)
	if got := ctx.Port(); got != "80" {
		t.Errorf("Port() = %q, want %q", got, "80")
	}
}

func TestContext_RemotePort(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	rec := httptest.NewRecorder()

	// Set RemoteAddr manually to simulate client port
	req.RemoteAddr = net.JoinHostPort("127.0.0.1", "56789")

	ctx := newContext(nil)
	ctx.init(req, rec)

	if got := ctx.RemotePort(); got != "56789" {
		t.Errorf("RemotePort() = %q, want %q", got, "56789")
	}
}

func TestContext_IP(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	rec := httptest.NewRecorder()

	req.RemoteAddr = net.JoinHostPort("192.168.1.50", "45678")

	ctx := newContext(nil)
	ctx.init(req, rec)

	if got := ctx.IP(); got != "192.168.1.50" {
		t.Errorf("IP() = %q, want %q", got, "192.168.1.50")
	}
}

func TestContext_Headers_GetSetAddDel(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	rec := httptest.NewRecorder()
	ctx := newContext(nil)
	ctx.init(req, rec)

	// Test setting header
	ctx.SetHeader("X-Test", "value1")
	if got := rec.Header().Get("X-Test"); got != "value1" {
		t.Errorf("SetHeader() = %q, want %q", got, "value1")
	}

	// Test adding header
	ctx.AddHeader("X-Test", "value2")
	values := rec.Header()["X-Test"]
	if len(values) != 2 || values[0] != "value1" || values[1] != "value2" {
		t.Errorf("AddHeader() = %v, want [value1 value2]", values)
	}

	// Test deleting header
	ctx.DelHeader("X-Test")
	if got := rec.Header().Get("X-Test"); got != "" {
		t.Errorf("DelHeader() = %q, want empty", got)
	}
}

func TestContext_GetHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	req.Header.Set("X-Custom", "abc123")
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	if got := ctx.GetHeader("X-Custom"); got != "abc123" {
		t.Errorf("GetHeader() = %q, want %q", got, "abc123")
	}
}

func TestContext_HeadersMap(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	req.Header.Set("X-One", "1")
	req.Header.Set("X-Two", "2")
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	h := ctx.Headers()
	if h.Get("X-One") != "1" || h.Get("X-Two") != "2" {
		t.Errorf("Headers() map = %v, missing expected values", h)
	}
}

func TestContext_SendHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	err := ctx.SendHeader("X-Custom-Header", "my-value")
	if err != nil {
		t.Errorf("SendHeader returned unexpected error: %v", err)
	}

	// verify header is set
	if got := rec.Header().Get("X-Custom-Header"); got != "my-value" {
		t.Errorf("expected X-Custom-Header to be %q, got %q", "my-value", got)
	}
}

func TestContext_SetContentType(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	ctx.SetContentType("application/json")

	// verify Content-Type is set
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("expected Content-Type to be %q, got %q", "application/json", got)
	}
}

func TestContext_RealIP(t *testing.T) {
	// --- Case 1: With X-Forwarded-For header ---
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 70.41.3.18, 150.172.238.178")
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	got := ctx.RealIP()
	want := "203.0.113.1" // first entry
	if got != want {
		t.Errorf("RealIP() with X-Forwarded-For = %q, want %q", got, want)
	}

	// --- Case 2: Without X-Forwarded-For header ---
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "192.168.1.100:12345"
	rec2 := httptest.NewRecorder()

	ctx2 := newContext(nil)
	ctx2.init(req2, rec2)

	got2 := ctx2.RealIP()
	want2 := "192.168.1.100" // comes from c.IP()
	if got2 != want2 {
		t.Errorf("RealIP() without header = %q, want %q", got2, want2)
	}
}

func TestContext_Body(t *testing.T) {
	bodyContent := `{"message":"hello"}`

	req := httptest.NewRequest("POST", "/submit", strings.NewReader(bodyContent))
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	got := ctx.Body()
	if string(got) != bodyContent {
		t.Errorf("Body() = %q, want %q", string(got), bodyContent)
	}
}

func TestContext_SendBytes_SendByte(t *testing.T) {
	// Create a test HTTP request and response recorder
	req := httptest.NewRequest("GET", "/bytes", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(nil)
	ctx.init(req, rec)

	// Test SendBytes
	data := []byte{0x01, 0x02, 0x03, 0x04}
	if err := ctx.SendBytes(data); err != nil {
		t.Fatalf("SendBytes failed: %v", err)
	}

	got := rec.Body.Bytes()
	if len(got) != len(data) {
		t.Fatalf("SendBytes wrote wrong length: got %d, want %d", len(got), len(data))
	}
	for i := range data {
		if got[i] != data[i] {
			t.Errorf("SendBytes byte %d = %v; want %v", i, got[i], data[i])
		}
	}

	// Reset recorder for SendByte test
	rec = httptest.NewRecorder()
	ctx.init(req, rec)

	// Test SendByte (single byte)
	single := byte(0x7F)
	if err := ctx.SendByte([]byte{single}); err != nil {
		t.Fatalf("SendByte failed: %v", err)
	}

	gotSingle := rec.Body.Bytes()
	if len(gotSingle) != 1 || gotSingle[0] != single {
		t.Errorf("SendByte wrote wrong value: got %v, want %v", gotSingle, []byte{single})
	}
}

func TestContext_BindJson_SendJson(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// --- Create a dummy Nexora instance with JSON encoder/decoder ---
	dummyNexora := &Nexora{
		JsonDecoder: func(data []byte, v any) error {
			return json.Unmarshal(data, v)
		},
		JsonEncoder: func(v any) ([]byte, error) {
			return json.Marshal(v)
		},
	}

	// --- Test successful BindJson ---
	reqBody := `{"name":"Alice","age":30}`
	req := httptest.NewRequest("POST", "/json", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	ctx := newContext(dummyNexora)
	ctx.init(req, rec)

	var p payload
	if err := ctx.BindJson(&p); err != nil {
		t.Fatalf("BindJson failed: %v", err)
	}
	if p.Name != "Alice" || p.Age != 30 {
		t.Errorf("BindJson parsed wrong values: %+v", p)
	}

	// --- Test BindJson with invalid JSON ---
	req3 := httptest.NewRequest("POST", "/json", strings.NewReader(`{"name":"Bob",`))
	req3.Header.Set("Content-Type", "application/json")
	rec3 := httptest.NewRecorder()

	ctx3 := newContext(dummyNexora)
	ctx3.init(req3, rec3)

	var p3 payload
	err := ctx3.BindJson(&p3)
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}

	// --- Test SendJson ---
	out := payload{Name: "Charlie", Age: 25}
	req4 := httptest.NewRequest("GET", "/json", nil)
	rec4 := httptest.NewRecorder()

	ctx4 := newContext(dummyNexora)
	ctx4.init(req4, rec4)

	if err := ctx4.SendJson(out); err != nil {
		t.Fatalf("SendJson failed: %v", err)
	}

	gotBody := rec4.Body.String()
	wantBody := `{"name":"Charlie","age":25}` // match struct tags
	if gotBody != wantBody {
		t.Errorf("SendJson wrote wrong body: got %q, want %q", gotBody, wantBody)
	}
	if ct := rec4.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("SendJson Content-Type = %q; want application/json", ct)
	}
}

func TestContext_SendSecureJson(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// --- Dummy Nexora instance ---
	dummyNexora := &Nexora{
		JsonDecoder: func(data []byte, v any) error {
			return json.Unmarshal(data, v)
		},
		JsonEncoder: func(v any) ([]byte, error) {
			return json.Marshal(v)
		},
		secureJsonPrefix: []byte("while(1);"),
	}

	// --- Test SendSecureJson ---
	out := payload{Name: "Daisy", Age: 40}
	req := httptest.NewRequest("GET", "/securejson", nil)
	rec := httptest.NewRecorder()

	ctx := newContext(dummyNexora)
	ctx.init(req, rec)

	if err := ctx.SendSecureJson(out); err != nil {
		t.Fatalf("SendSecureJson failed: %v", err)
	}

	gotBody := rec.Body.String()
	wantBody := `while(1);{"name":"Daisy","age":40}`
	if gotBody != wantBody {
		t.Errorf("SendSecureJson wrote wrong body: got %q, want %q", gotBody, wantBody)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("SendSecureJson Content-Type = %q; want application/json", ct)
	}
}

func TestContext_BindXml_SendXml(t *testing.T) {
	type payload struct {
		Name string `xml:"name"`
		Age  int    `xml:"age"`
	}

	// --- Create a dummy Nexora instance with XML encoder/decoder ---
	dummyNexora := &Nexora{
		XmlDecoder: func(data []byte, v any) error {
			return xml.Unmarshal(data, v)
		},
		XmlEncoder: func(v any) ([]byte, error) {
			return xml.Marshal(v)
		},
		XmlIndentationEncoder: func(v any, prefix, indent string) ([]byte, error) {
			return xml.MarshalIndent(v, prefix, indent)
		},
	}

	// --- Test successful BindXml ---
	reqBody := `<payload><name>Alice</name><age>30</age></payload>`
	req := httptest.NewRequest("POST", "/xml", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/xml")
	rec := httptest.NewRecorder()

	ctx := newContext(dummyNexora)
	ctx.init(req, rec)

	var p payload
	if err := ctx.BindXml(&p); err != nil {
		t.Fatalf("BindXml failed: %v", err)
	}
	if p.Name != "Alice" || p.Age != 30 {
		t.Errorf("BindXml parsed wrong values: %+v", p)
	}

	// --- Test BindXml with invalid XML ---
	req2 := httptest.NewRequest("POST", "/xml", strings.NewReader(`<payload><name>Bob</age>`))
	req2.Header.Set("Content-Type", "application/xml")
	rec2 := httptest.NewRecorder()

	ctx2 := newContext(dummyNexora)
	ctx2.init(req2, rec2)

	var p2 payload
	err := ctx2.BindXml(&p2)
	if err == nil {
		t.Fatal("expected error for malformed XML, got nil")
	}

	// --- Test SendXml ---
	out := payload{Name: "Charlie", Age: 25}
	req3 := httptest.NewRequest("GET", "/xml", nil)
	rec3 := httptest.NewRecorder()

	ctx3 := newContext(dummyNexora)
	ctx3.init(req3, rec3)

	if err := ctx3.SendXml(out); err != nil {
		t.Fatalf("SendXml failed: %v", err)
	}

	gotBody := rec3.Body.String()
	wantBody := `<payload><name>Charlie</name><age>25</age></payload>`
	if gotBody != wantBody {
		t.Errorf("SendXml wrote wrong body:\n got  %q\n want %q", gotBody, wantBody)
	}
	if ct := rec3.Header().Get("Content-Type"); ct != "application/xml" {
		t.Errorf("SendXml Content-Type = %q; want application/xml", ct)
	}

	// --- Test SendPrettyXml ---
	req4 := httptest.NewRequest("GET", "/xml", nil)
	rec4 := httptest.NewRecorder()

	ctx4 := newContext(dummyNexora)
	ctx4.init(req4, rec4)

	if err := ctx4.SendPrettyXml(out); err != nil {
		t.Fatalf("SendPrettyXml failed: %v", err)
	}
	if !strings.Contains(rec4.Body.String(), "\n") {
		t.Errorf("SendPrettyXml did not pretty-print XML, got: %q", rec4.Body.String())
	}

	// --- Test SendIndentXml with custom indent ---
	req5 := httptest.NewRequest("GET", "/xml", nil)
	rec5 := httptest.NewRecorder()

	ctx5 := newContext(dummyNexora)
	ctx5.init(req5, rec5)

	if err := ctx5.SendIndentXml(out, "", "\t"); err != nil {
		t.Fatalf("SendIndentXml failed: %v", err)
	}
	if !strings.Contains(rec5.Body.String(), "\t<name>") {
		t.Errorf("SendIndentXml did not apply custom indent, got: %q", rec5.Body.String())
	}
}
