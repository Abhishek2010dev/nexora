// Copyright (c) 2025 Abhishek Kumar Singh.
// Copyright (c) 2013 Julien Schmidt
// Copyright (c) 2015-2016, 招牌疯子
// Copyright (c) 2018-present Sergio Andres Virviescas Santana, fasthttp
// All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file at https://raw.githubusercontent.com/julienschmidt/httprouter/master/LICENSE.
package nexora

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	gstrings "github.com/savsgio/gotils/strings"
	"github.com/valyala/bytebufferpool"
)

const (
	errSetHandler         = "nexora: a handler is already registered for path '%s'"
	errSetWildcardHandler = "nexora: a wildcard handler is already registered for path '%s'"
	errWildPathConflict   = "nexora: '%s' in new path '%s' conflicts with existing wild path '%s' in existing prefix '%s'"
	errWildcardConflict   = "nexora: '%s' in new path '%s' conflicts with existing wildcard '%s' in existing prefix '%s'"
	errWildcardSlash      = "nexora: no / before wildcard in path '%s'"
	errWildcardNotAtEnd   = "nexora: wildcard routes are only allowed at the end of the path in path '%s'"
)

type radixError struct {
	msg    string
	params []any
}

func (err radixError) Error() string {
	return fmt.Sprintf(err.msg, err.params...)
}

func newRadixError(msg string, params ...any) radixError {
	return radixError{msg, params}
}

type nodeType uint8

const (
	root nodeType = iota
	static
	wildcard
	param
)

type nodeWildcard struct {
	path     string
	paramKey string
	handlers []Handler
}

type node struct {
	nType nodeType

	path         string
	tsr          bool
	handlers     []Handler
	hasWildChild bool
	children     []*node
	wildcard     *nodeWildcard

	paramKeys  []string
	paramRegex *regexp.Regexp
}

type wildPath struct {
	path  string
	keys  []string
	start int
	end   int
	pType nodeType

	pattern string
	regex   *regexp.Regexp
}

func newNode(path string) *node {
	return &node{
		nType: static,
		path:  path,
	}
}

// conflict raises a panic with some details
func (n *nodeWildcard) conflict(path, fullPath string) error {
	prefix := fullPath[:strings.LastIndex(fullPath, path)] + n.path

	return newRadixError(errWildcardConflict, path, fullPath, n.path, prefix)
}

// wildPathConflict raises a panic with some details
func (n *node) wildPathConflict(path, fullPath string) error {
	pathSeg := strings.SplitN(path, "/", 2)[0]
	prefix := fullPath[:strings.LastIndex(fullPath, path)] + n.path

	return newRadixError(errWildPathConflict, pathSeg, fullPath, n.path, prefix)
}

// clone clones the current node in a new pointer
func (n node) clone() *node {
	cloneNode := new(node)
	cloneNode.nType = n.nType
	cloneNode.path = n.path
	cloneNode.tsr = n.tsr
	cloneNode.handlers = n.handlers

	if len(n.children) > 0 {
		cloneNode.children = make([]*node, len(n.children))

		for i, child := range n.children {
			cloneNode.children[i] = child.clone()
		}
	}

	if n.wildcard != nil {
		cloneNode.wildcard = &nodeWildcard{
			path:     n.wildcard.path,
			paramKey: n.wildcard.paramKey,
			handlers: n.wildcard.handlers,
		}
	}

	if len(n.paramKeys) > 0 {
		cloneNode.paramKeys = make([]string, len(n.paramKeys))
		copy(cloneNode.paramKeys, n.paramKeys)
	}

	cloneNode.paramRegex = n.paramRegex

	return cloneNode
}

func (n *node) split(i int) {
	cloneChild := n.clone()
	cloneChild.nType = static
	cloneChild.path = cloneChild.path[i:]
	cloneChild.paramKeys = nil
	cloneChild.paramRegex = nil

	n.path = n.path[:i]
	n.handlers = nil
	n.tsr = false
	n.wildcard = nil
	n.children = append(n.children[:0], cloneChild)
}

func (n *node) findEndIndexAndValues(path string) (int, []string) {
	index := n.paramRegex.FindStringSubmatchIndex(path)
	if len(index) == 0 || index[0] != 0 {
		return -1, nil
	}

	end := index[1]

	index = index[2:]
	values := make([]string, len(index)/2)

	i := 0
	for j := range index {
		if (j+1)%2 != 0 {
			continue
		}

		values[i] = gstrings.Copy(path[index[j-1]:index[j]])

		i++
	}

	return end, values
}

func (n *node) setHandler(handlers []Handler, fullPath string) (*node, error) {
	if n.handlers != nil || n.tsr {
		return n, newRadixError(errSetHandler, fullPath)
	}

	n.handlers = handlers
	foundTSR := false

	// Set TSR in method
	for i := range n.children {
		child := n.children[i]

		if child.path != "/" {
			continue
		}

		child.tsr = true
		foundTSR = true

		break
	}

	if n.path != "/" && !foundTSR {
		if strings.HasSuffix(n.path, "/") {
			n.split(len(n.path) - 1)
			n.tsr = true
		} else {
			childTSR := newNode("/")
			childTSR.tsr = true
			n.children = append(n.children, childTSR)
		}
	}

	return n, nil
}

func (n *node) insert(path, fullPath string, handlers []Handler) (*node, error) {
	end := segmentEndIndex(path, true)
	child := newNode(path)

	wp := findWildPath(path, fullPath)
	if wp != nil {
		j := end
		if wp.start > 0 {
			j = wp.start
		}

		child.path = path[:j]

		if wp.start > 0 {
			n.children = append(n.children, child)

			return child.insert(path[j:], fullPath, handlers)
		}

		switch wp.pType {
		case param:
			n.hasWildChild = true

			child.nType = wp.pType
			child.paramKeys = wp.keys
			child.paramRegex = wp.regex
		case wildcard:
			if len(path) == end && n.path[len(n.path)-1] != '/' {
				return nil, newRadixError(errWildcardSlash, fullPath)
			} else if len(path) != end {
				return nil, newRadixError(errWildcardNotAtEnd, fullPath)
			}

			if n.path != "/" && n.path[len(n.path)-1] == '/' {
				n.split(len(n.path) - 1)
				n.tsr = true

				n = n.children[0]
			}

			if n.wildcard != nil {
				if n.wildcard.path == path {
					return n, newRadixError(errSetWildcardHandler, fullPath)
				}

				return nil, n.wildcard.conflict(path, fullPath)
			}

			n.wildcard = &nodeWildcard{
				path:     wp.path,
				paramKey: wp.keys[0],
				handlers: handlers,
			}

			return n, nil
		}

		path = path[wp.end:]

		if len(path) > 0 {
			n.children = append(n.children, child)

			return child.insert(path, fullPath, handlers)
		}
	}

	child.handlers = handlers
	n.children = append(n.children, child)

	if child.path == "/" {
		// Add TSR when split a edge and the remain path to insert is "/"
		n.tsr = true
	} else if strings.HasSuffix(child.path, "/") {
		child.split(len(child.path) - 1)
		child.tsr = true
	} else {
		childTSR := newNode("/")
		childTSR.tsr = true
		child.children = append(child.children, childTSR)
	}

	return child, nil
}

// add adds the handler to node for the given path
func (n *node) add(path, fullPath string, handlers []Handler) (*node, error) {
	if len(path) == 0 {
		return n.setHandler(handlers, fullPath)
	}

	for _, child := range n.children {
		i := longestCommonPrefix(path, child.path)
		if i == 0 {
			continue
		}

		switch child.nType {
		case static:
			if len(child.path) > i {
				child.split(i)
			}

			if len(path) > i {
				return child.add(path[i:], fullPath, handlers)
			}
		case param:
			wp := findWildPath(path, fullPath)

			isParam := wp.start == 0 && wp.pType == param
			hasHandler := child.handlers != nil || handlers == nil

			if len(path) == wp.end && isParam && hasHandler {
				// The current segment is a param and it's duplicated
				if child.path == path {
					return child, newRadixError(errSetHandler, fullPath)
				}

				return nil, child.wildPathConflict(path, fullPath)
			}

			if len(path) > i {
				if child.path == wp.path {
					return child.add(path[i:], fullPath, handlers)
				}

				return n.insert(path, fullPath, handlers)
			}
		}

		if path == "/" {
			n.tsr = true
		}

		return child.setHandler(handlers, fullPath)
	}

	return n.insert(path, fullPath, handlers)
}

func (n *node) getFromChild(path string) ([]Handler, map[string]string, bool) {
	for _, child := range n.children {
		switch child.nType {
		case static:
			if path[0] != child.path[0] {
				continue
			}
			if len(path) > len(child.path) {
				if path[:len(child.path)] != child.path {
					continue
				}
				h, params, tsr := child.getFromChild(path[len(child.path):])
				if h != nil || tsr {
					return h, params, tsr
				}
			} else if path == child.path {
				switch {
				case child.tsr:
					return nil, nil, true
				case child.handlers != nil:
					return child.handlers, nil, false
				case child.wildcard != nil:
					params := map[string]string{
						child.wildcard.paramKey: "",
					}
					return child.wildcard.handlers, params, false
				}
				return nil, nil, false
			}

		case param:
			end := segmentEndIndex(path, false)
			paramVal := path[:end]
			values := []string{gstrings.Copy(paramVal)}

			if child.paramRegex != nil {
				end, values = child.findEndIndexAndValues(paramVal)
				if end == -1 {
					continue
				}
			}

			if len(path) > end {
				h, params, tsr := child.getFromChild(path[end:])
				if h != nil {
					if params == nil {
						params = make(map[string]string)
					}
					for i, key := range child.paramKeys {
						params[key] = values[i]
					}
					return h, params, false
				} else if tsr {
					return nil, nil, true
				}
			} else if len(path) == end {
				if child.handlers != nil {
					params := make(map[string]string)
					for i, key := range child.paramKeys {
						params[key] = values[i]
					}
					return child.handlers, params, false
				}
				if child.tsr {
					return nil, nil, true
				}
				// Try another child
				continue
			}

		default:
			panic("invalid node type")
		}
	}

	if n.wildcard != nil {
		params := map[string]string{
			n.wildcard.paramKey: gstrings.Copy(path),
		}
		return n.wildcard.handlers, params, false
	}

	return nil, nil, false
}

func (n *node) find(path string, buf *bytebufferpool.ByteBuffer) (bool, bool) {
	if len(path) > len(n.path) {
		if !strings.EqualFold(path[:len(n.path)], n.path) {
			return false, false
		}

		path = path[len(n.path):]
		buf.WriteString(n.path)

		found, tsr := n.findFromChild(path, buf)
		if found {
			return found, tsr
		}

		bufferRemoveString(buf, n.path)

	} else if strings.EqualFold(path, n.path) {
		buf.WriteString(n.path)

		if n.tsr {
			if n.path == "/" {
				bufferRemoveString(buf, n.path)
			} else {
				buf.WriteByte('/')
			}

			return true, true
		}

		if n.handlers != nil {
			return true, false
		} else {
			bufferRemoveString(buf, n.path)
		}
	}

	return false, false
}

func (n *node) findFromChild(path string, buf *bytebufferpool.ByteBuffer) (bool, bool) {
	for _, child := range n.children {
		switch child.nType {
		case static:
			found, tsr := child.find(path, buf)
			if found {
				return found, tsr
			}

		case param:
			end := segmentEndIndex(path, false)

			if child.paramRegex != nil {
				end, _ = child.findEndIndexAndValues(path[:end])
				if end == -1 {
					continue
				}
			}

			buf.WriteString(path[:end])

			if len(path) > end {
				found, tsr := child.findFromChild(path[end:], buf)
				if found {
					return found, tsr
				}

			} else if len(path) == end {
				if child.tsr {
					buf.WriteByte('/')

					return true, true
				}

				if child.handlers != nil {
					return true, false
				}
			}

			bufferRemoveString(buf, path[:end])

		default:
			panic("invalid node type")
		}
	}

	if n.wildcard != nil {
		buf.WriteString(path)

		return true, false
	}

	return false, false
}

// sort sorts the current node and their children
func (n *node) sort() {
	for _, child := range n.children {
		child.sort()
	}

	sort.Sort(n)
}

// Len returns the total number of children the node has
func (n *node) Len() int {
	return len(n.children)
}

// Swap swaps the order of children nodes
func (n *node) Swap(i, j int) {
	n.children[i], n.children[j] = n.children[j], n.children[i]
}

// Less checks if the node 'i' has less priority than the node 'j'
func (n *node) Less(i, j int) bool {
	if n.children[i].nType < n.children[j].nType {
		return true
	} else if n.children[i].nType > n.children[j].nType {
		return false
	}

	return len(n.children[i].children) > len(n.children[j].children)
}

// tree is a routes storage
type tree struct {
	root *node

	// If enabled, the node handler could be updated
	Mutable bool
}

// New returns an empty routes storage
func newTree() *tree {
	return &tree{
		root: &node{
			nType: root,
		},
	}
}

// Add adds a node with the given handle to the path.
//
// WARNING: Not concurrency-safe!
func (t *tree) Add(path string, handlers []Handler) {
	if !strings.HasPrefix(path, "/") {
		panicf("path must begin with '/' in path '%s'", path)
	} else if handlers == nil {
		panic("nil handler")
	}

	fullPath := path

	i := longestCommonPrefix(path, t.root.path)
	if i > 0 {
		if len(t.root.path) > i {
			t.root.split(i)
		}

		path = path[i:]
	}

	n, err := t.root.add(path, fullPath, handlers)
	if err != nil {
		var radixErr radixError

		if errors.As(err, &radixErr) && t.Mutable && !n.tsr {
			switch radixErr.msg {
			case errSetHandler:
				n.handlers = handlers
				return
			case errSetWildcardHandler:
				n.wildcard.handlers = handlers
				return
			}
		}

		panic(err)
	}

	if len(t.root.path) == 0 {
		t.root = t.root.children[0]
		t.root.nType = root
	}

	// Reorder the nodes
	t.root.sort()
}

// Get returns the handler(s) registered with the given path.
// It also returns any route parameters as map[string]string and a bool indicating a TSR (trailing slash redirect).
func (t *tree) Get(path string) ([]Handler, map[string]string, bool) {
	if len(path) > len(t.root.path) {
		if path[:len(t.root.path)] != t.root.path {
			return nil, nil, false
		}

		path = path[len(t.root.path):]
		return t.root.getFromChild(path)

	} else if path == t.root.path {
		switch {
		case t.root.tsr:
			return nil, nil, true
		case t.root.handlers != nil:
			return t.root.handlers, nil, false
		case t.root.wildcard != nil:
			params := map[string]string{
				t.root.wildcard.paramKey: "",
			}
			return t.root.wildcard.handlers, params, false
		}
	}

	return nil, nil, false
}

// FindCaseInsensitivePath makes a case-insensitive lookup of the given path
// and tries to find a handler.
// It can optionally also fix trailing slashes.
// It returns the case-corrected path and a bool indicating whether the lookup
// was successful.
func (t *tree) FindCaseInsensitivePath(path string, fixTrailingSlash bool, buf *bytebufferpool.ByteBuffer) bool {
	found, tsr := t.root.find(path, buf)

	if !found || (tsr && !fixTrailingSlash) {
		buf.Reset()

		return false
	}

	return true
}
