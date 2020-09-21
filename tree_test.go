// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/julienschmidt/httprouter/blob/master/LICENSE.

package clevergo

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Used as a workaround since we can't compare functions or their addresses
var fakeHandlerValue string

func fakeHandler(val string) Handle {
	return func(c *Context) error {
		fakeHandlerValue = val
		return nil
	}
}

type testRequests []struct {
	path       string
	nilHandler bool
	route      string
	ps         Params
}

func checkRequests(t *testing.T, tree *node, requests testRequests) {
	for _, request := range requests {
		ps := make(Params, 0, 20)
		handler, _ := tree.getValue(request.path, &ps, false)

		if request.nilHandler {
			assert.Nil(t, handler, "handle mismatch for route '%s': Expected nil handle", request.path)
		} else {
			assert.NotNil(t, handler, "handle mismatch for route '%s': Expected non-nil handle", request.path)
			handler.handle(newContext(nil, nil))
			assert.Equalf(t, request.route, fakeHandlerValue, "handle mismatch for route '%s'", request.path)
		}

		if request.ps == nil {
			assert.Len(t, ps, 0)
		} else {
			assert.Equal(t, request.ps, ps)
		}
	}
}

func checkPriorities(t *testing.T, n *node) uint32 {
	var prio uint32
	for i := range n.children {
		prio += checkPriorities(t, n.children[i])
	}

	if n.route != nil {
		prio++
	}

	assert.Equalf(t, prio, n.priority, "priority mismatch for node '%s'", n.path)

	return prio
}

func TestCountParams(t *testing.T) {
	if countParams("/path/:param1/static/*catch-all") != 2 {
		t.Fail()
	}
	if countParams(strings.Repeat("/:param", 256)) != 256 {
		t.Fail()
	}
}

func TestTreeAddAndGet(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/hi",
		"/contact",
		"/co",
		"/c",
		"/a",
		"/ab",
		"/doc/",
		"/doc/go_faq.html",
		"/doc/go1.html",
		"/α",
		"/β",
	}
	for _, route := range routes {
		tree.addRoute(route, newRoute(route, fakeHandler(route)))
	}

	checkRequests(t, tree, testRequests{
		{"/a", false, "/a", nil},
		{"/", true, "", nil},
		{"/hi", false, "/hi", nil},
		{"/contact", false, "/contact", nil},
		{"/co", false, "/co", nil},
		{"/con", true, "", nil},  // key mismatch
		{"/cona", true, "", nil}, // key mismatch
		{"/no", true, "", nil},   // no matching child
		{"/ab", false, "/ab", nil},
		{"/α", false, "/α", nil},
		{"/β", false, "/β", nil},
	})

	checkPriorities(t, tree)
}

func TestTreeWildcard(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/",
		"/cmd/:tool/:sub",
		"/cmd/:tool/",
		"/src/*filepath",
		"/search/",
		"/search/:query",
		"/user_:name",
		"/user_:name/about",
		"/files/:dir/*filepath",
		"/doc/",
		"/doc/go_faq.html",
		"/doc/go1.html",
		"/info/:user/public",
		"/info/:user/project/:project",
	}
	for _, route := range routes {
		tree.addRoute(route, newRoute(route, fakeHandler(route)))
	}

	checkRequests(t, tree, testRequests{
		{"/", false, "/", nil},
		{"/cmd/test/", false, "/cmd/:tool/", Params{Param{"tool", "test"}}},
		{"/cmd/test", true, "", Params{Param{"tool", "test"}}},
		{"/cmd/test/3", false, "/cmd/:tool/:sub", Params{Param{"tool", "test"}, Param{"sub", "3"}}},
		{"/src/", false, "/src/*filepath", Params{Param{"filepath", "/"}}},
		{"/src/some/file.png", false, "/src/*filepath", Params{Param{"filepath", "/some/file.png"}}},
		{"/search/", false, "/search/", nil},
		{"/search/someth!ng+in+ünìcodé", false, "/search/:query", Params{Param{"query", "someth!ng+in+ünìcodé"}}},
		{"/search/someth!ng+in+ünìcodé/", true, "", Params{Param{"query", "someth!ng+in+ünìcodé"}}},
		{"/user_gopher", false, "/user_:name", Params{Param{"name", "gopher"}}},
		{"/user_gopher/about", false, "/user_:name/about", Params{Param{"name", "gopher"}}},
		{"/files/js/inc/framework.js", false, "/files/:dir/*filepath", Params{Param{"dir", "js"}, Param{"filepath", "/inc/framework.js"}}},
		{"/info/gordon/public", false, "/info/:user/public", Params{Param{"user", "gordon"}}},
		{"/info/gordon/project/go", false, "/info/:user/project/:project", Params{Param{"user", "gordon"}, Param{"project", "go"}}},
	})

	checkPriorities(t, tree)
}

func catchPanic(testFunc func()) (recv interface{}) {
	defer func() {
		recv = recover()
	}()

	testFunc()
	return
}

type testRoute struct {
	*Route
	conflict bool
}

func newTestRoute(path string, conflict bool) testRoute {
	return testRoute{
		Route:    newRoute(path, nil),
		conflict: conflict,
	}
}

func testRoutes(t *testing.T, routes []testRoute) {
	tree := &node{}

	for _, route := range routes {
		recv := catchPanic(func() {
			tree.addRoute(route.path, route.Route)
		})

		if route.conflict {
			assert.NotNilf(t, recv, "no panic for conflicting route '%s'", route.path)
		} else {
			assert.Nilf(t, recv, "unexpected panic for route '%s': %v", route.path, recv)
		}
	}
}

func TestTreeWildcardConflict(t *testing.T) {
	routes := []testRoute{
		newTestRoute("/cmd/:tool/:sub", false),
		newTestRoute("/cmd/vet", true),
		newTestRoute("/src/*filepath", false),
		newTestRoute("/src/*filepathx", true),
		newTestRoute("/src/", true),
		newTestRoute("/src1/", false),
		newTestRoute("/src1/*filepath", true),
		newTestRoute("/src2*filepath", true),
		newTestRoute("/search/:query", false),
		newTestRoute("/search/invalid", true),
		newTestRoute("/user_:name", false),
		newTestRoute("/user_x", true),
		newTestRoute("/user_:name", true),
		newTestRoute("/id:id", false),
		newTestRoute("/id/:id", true),
	}
	testRoutes(t, routes)
}

func TestTreeChildConflict(t *testing.T) {
	routes := []testRoute{
		newTestRoute("/cmd/vet", false),
		newTestRoute("/cmd/:tool/:sub", true),
		newTestRoute("/src/AUTHORS", false),
		newTestRoute("/src/*filepath", true),
		newTestRoute("/user_x", false),
		newTestRoute("/user_:name", true),
		newTestRoute("/id/:id", false),
		newTestRoute("/id:id", true),
		newTestRoute("/:id", true),
		newTestRoute("/*filepath", true),
	}
	testRoutes(t, routes)
}

func TestTreeDupliatePath(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/",
		"/doc/",
		"/src/*filepath",
		"/search/:query",
		"/user_:name",
	}
	for _, route := range routes {
		recv := catchPanic(func() {
			tree.addRoute(route, newRoute(route, fakeHandler(route)))
		})
		assert.Nilf(t, recv, "panic inserting route '%s': %v", route, recv)

		// Add again
		recv = catchPanic(func() {
			tree.addRoute(route, nil)
		})
		assert.NotNilf(t, recv, "no panic while inserting duplicate route '%s", route)
	}

	checkRequests(t, tree, testRequests{
		{"/", false, "/", nil},
		{"/doc/", false, "/doc/", nil},
		{"/src/some/file.png", false, "/src/*filepath", Params{Param{"filepath", "/some/file.png"}}},
		{"/search/someth!ng+in+ünìcodé", false, "/search/:query", Params{Param{"query", "someth!ng+in+ünìcodé"}}},
		{"/user_gopher", false, "/user_:name", Params{Param{"name", "gopher"}}},
	})
}

func TestEmptyWildcardName(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/user:",
		"/user:/",
		"/cmd/:/",
		"/src/*",
	}
	for _, route := range routes {
		recv := catchPanic(func() {
			tree.addRoute(route, nil)
		})
		assert.NotNilf(t, recv, "no panic while inserting route with empty wildcard name '%s", route)
	}
}

func TestTreeCatchAllConflict(t *testing.T) {
	routes := []testRoute{
		newTestRoute("/src/*filepath/x", true),
		newTestRoute("/src2/", false),
		newTestRoute("/src2/*filepath/x", true),
		newTestRoute("/src3/*filepath", false),
		newTestRoute("/src3/*filepath/x", true),
	}
	testRoutes(t, routes)
}

func TestTreeCatchAllConflictRoot(t *testing.T) {
	routes := []testRoute{
		newTestRoute("/", false),
		newTestRoute("/*filepath", true),
	}
	testRoutes(t, routes)
}

func TestTreeCatchMaxParams(t *testing.T) {
	tree := &node{}
	var route = "/cmd/*filepath"
	tree.addRoute(route, newRoute(route, fakeHandler(route)))
}

func TestTreeDoubleWildcard(t *testing.T) {
	const panicMsg = "only one wildcard per path segment is allowed"

	routes := [...]string{
		"/:foo:bar",
		"/:foo:bar/",
		"/:foo*bar",
	}

	for _, route := range routes {
		tree := &node{}
		recv := catchPanic(func() {
			tree.addRoute(route, newRoute(route, nil))
		})

		rs, ok := recv.(string)
		assert.True(t, ok)
		assert.Truef(t, strings.HasPrefix(rs, panicMsg), `"Expected panic "%s" for route '%s', got "%v"`, panicMsg, route, recv)
	}
}

/*func TestTreeDuplicateWildcard(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/:id/:name/:id",
	}
	for _, route := range routes {
		...
	}
}*/

func TestTreeTrailingSlashRedirect(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/hi",
		"/b/",
		"/search/:query",
		"/cmd/:tool/",
		"/src/*filepath",
		"/x",
		"/x/y",
		"/y/",
		"/y/z",
		"/0/:id",
		"/0/:id/1",
		"/1/:id/",
		"/1/:id/2",
		"/aa",
		"/a/",
		"/admin",
		"/admin/:category",
		"/admin/:category/:page",
		"/doc",
		"/doc/go_faq.html",
		"/doc/go1.html",
		"/no/a",
		"/no/b",
		"/api/hello/:name",
		"/vendor/:x/*y",
	}
	for _, route := range routes {
		recv := catchPanic(func() {
			tree.addRoute(route, newRoute(route, fakeHandler(route)))
		})
		assert.Nilf(t, recv, "panic inserting route '%s'", route)
	}

	tsrRoutes := [...]string{
		"/hi/",
		"/b",
		"/search/gopher/",
		"/cmd/vet",
		"/src",
		"/x/",
		"/y",
		"/0/go/",
		"/1/go",
		"/a",
		"/admin/",
		"/admin/config/",
		"/admin/config/permissions/",
		"/doc/",
		"/vendor/x",
	}
	for _, route := range tsrRoutes {
		handler, tsr := tree.getValue(route, nil, false)
		assert.Nilf(t, handler, "non-nil handler for TSR route '%s", route)
		assert.Truef(t, tsr, "expected TSR recommendation for route '%s'", route)
	}

	noTsrRoutes := [...]string{
		"/",
		"/no",
		"/no/",
		"/_",
		"/_/",
		"/api/world/abc",
	}
	for _, route := range noTsrRoutes {
		handler, tsr := tree.getValue(route, nil, false)
		assert.Nilf(t, handler, "non-nil handler for No-TSR route '%s", route)
		assert.Falsef(t, tsr, "expected no TSR recommendation for route '%s'", route)
	}
}

func TestTreeRootTrailingSlashRedirect(t *testing.T) {
	tree := &node{}

	recv := catchPanic(func() {
		tree.addRoute("/:test", newRoute("/:test", fakeHandler("/:test")))
	})
	assert.Nilf(t, recv, "panic inserting test route: %v", recv)

	handler, tsr := tree.getValue("/", nil, false)
	assert.Nil(t, handler)
	assert.False(t, tsr)
}

func TestTreeFindCaseInsensitivePath(t *testing.T) {
	tree := &node{}

	longPath := "/l" + strings.Repeat("o", 128) + "ng"
	lOngPath := "/l" + strings.Repeat("O", 128) + "ng/"

	routes := [...]string{
		"/hi",
		"/b/",
		"/ABC/",
		"/search/:query",
		"/cmd/:tool/",
		"/src/*filepath",
		"/x",
		"/x/y",
		"/y/",
		"/y/z",
		"/0/:id",
		"/0/:id/1",
		"/1/:id/",
		"/1/:id/2",
		"/aa",
		"/a/",
		"/doc",
		"/doc/go_faq.html",
		"/doc/go1.html",
		"/doc/go/away",
		"/no/a",
		"/no/b",
		"/Π",
		"/u/apfêl/",
		"/u/äpfêl/",
		"/u/öpfêl",
		"/v/Äpfêl/",
		"/v/Öpfêl",
		"/w/♬",  // 3 byte
		"/w/♭/", // 3 byte, last byte differs
		"/w/𠜎",  // 4 byte
		"/w/𠜏/", // 4 byte
		longPath,
	}

	for _, route := range routes {
		recv := catchPanic(func() {
			tree.addRoute(route, newRoute(route, fakeHandler(route)))
		})
		assert.Nilf(t, recv, "panic inserting route '%s': %v", route, recv)
	}

	// Check out == in for all registered routes
	// With fixTrailingSlash = true
	for _, route := range routes {
		out, found := tree.findCaseInsensitivePath(route, true)
		assert.Truef(t, found, "Route '%s' not found!", route)
		assert.Equalf(t, route, out, "Wrong result for route '%s'", route)
	}
	// With fixTrailingSlash = false
	for _, route := range routes {
		out, found := tree.findCaseInsensitivePath(route, false)
		assert.Truef(t, found, "Route '%s' not found!", route)
		assert.Equalf(t, route, out, "Wrong result for route '%s'", route)
	}

	tests := []struct {
		in    string
		out   string
		found bool
		slash bool
	}{
		{"/HI", "/hi", true, false},
		{"/HI/", "/hi", true, true},
		{"/B", "/b/", true, true},
		{"/B/", "/b/", true, false},
		{"/abc", "/ABC/", true, true},
		{"/abc/", "/ABC/", true, false},
		{"/aBc", "/ABC/", true, true},
		{"/aBc/", "/ABC/", true, false},
		{"/abC", "/ABC/", true, true},
		{"/abC/", "/ABC/", true, false},
		{"/SEARCH/QUERY", "/search/QUERY", true, false},
		{"/SEARCH/QUERY/", "/search/QUERY", true, true},
		{"/CMD/TOOL/", "/cmd/TOOL/", true, false},
		{"/CMD/TOOL", "/cmd/TOOL/", true, true},
		{"/SRC/FILE/PATH", "/src/FILE/PATH", true, false},
		{"/x/Y", "/x/y", true, false},
		{"/x/Y/", "/x/y", true, true},
		{"/X/y", "/x/y", true, false},
		{"/X/y/", "/x/y", true, true},
		{"/X/Y", "/x/y", true, false},
		{"/X/Y/", "/x/y", true, true},
		{"/Y/", "/y/", true, false},
		{"/Y", "/y/", true, true},
		{"/Y/z", "/y/z", true, false},
		{"/Y/z/", "/y/z", true, true},
		{"/Y/Z", "/y/z", true, false},
		{"/Y/Z/", "/y/z", true, true},
		{"/y/Z", "/y/z", true, false},
		{"/y/Z/", "/y/z", true, true},
		{"/Aa", "/aa", true, false},
		{"/Aa/", "/aa", true, true},
		{"/AA", "/aa", true, false},
		{"/AA/", "/aa", true, true},
		{"/aA", "/aa", true, false},
		{"/aA/", "/aa", true, true},
		{"/A/", "/a/", true, false},
		{"/A", "/a/", true, true},
		{"/DOC", "/doc", true, false},
		{"/DOC/", "/doc", true, true},
		{"/NO", "", false, true},
		{"/DOC/GO", "", false, true},
		{"/π", "/Π", true, false},
		{"/π/", "/Π", true, true},
		{"/u/ÄPFÊL/", "/u/äpfêl/", true, false},
		{"/u/ÄPFÊL", "/u/äpfêl/", true, true},
		{"/u/ÖPFÊL/", "/u/öpfêl", true, true},
		{"/u/ÖPFÊL", "/u/öpfêl", true, false},
		{"/v/äpfêL/", "/v/Äpfêl/", true, false},
		{"/v/äpfêL", "/v/Äpfêl/", true, true},
		{"/v/öpfêL/", "/v/Öpfêl", true, true},
		{"/v/öpfêL", "/v/Öpfêl", true, false},
		{"/w/♬/", "/w/♬", true, true},
		{"/w/♭", "/w/♭/", true, true},
		{"/w/𠜎/", "/w/𠜎", true, true},
		{"/w/𠜏", "/w/𠜏/", true, true},
		{lOngPath, longPath, true, true},
	}
	// With fixTrailingSlash = true
	for _, test := range tests {
		out, found := tree.findCaseInsensitivePath(test.in, true)
		assert.Equal(t, test.found, found)
		if found {
			assert.Equal(t, test.out, out)
		}
	}
	// With fixTrailingSlash = false
	for _, test := range tests {
		out, found := tree.findCaseInsensitivePath(test.in, false)
		if test.slash {
			// test needs a trailingSlash fix. It must not be found!
			assert.Falsef(t, found, "Found without fixTrailingSlash: %s; got %s", test.in, out)
		} else {
			assert.Equal(t, test.found, found)
			if found {
				assert.Equal(t, test.out, out)
			}
		}
	}
}

func TestTreeInvalidNodeType(t *testing.T) {
	const panicMsg = "invalid node type"

	tree := &node{}
	tree.addRoute("/", newRoute("/", fakeHandler("/")))
	tree.addRoute("/:page", newRoute("/:page", fakeHandler("/:page")))

	// set invalid node type
	tree.children[0].nType = 42

	// normal lookup
	recv := catchPanic(func() {
		tree.getValue("/test", nil, false)
	})
	rs, ok := recv.(string)
	assert.True(t, ok)
	assert.Equal(t, panicMsg, rs)

	// case-insensitive lookup
	recv = catchPanic(func() {
		tree.findCaseInsensitivePath("/test", true)
	})
	rs, ok = recv.(string)
	assert.True(t, ok)
	assert.Equal(t, panicMsg, rs)
}

func TestTreeWildcardConflictEx(t *testing.T) {
	conflicts := [...]struct {
		route        string
		segPath      string
		existPath    string
		existSegPath string
	}{
		{"/who/are/foo", "/foo", `/who/are/\*you`, `/\*you`},
		{"/who/are/foo/", "/foo/", `/who/are/\*you`, `/\*you`},
		{"/who/are/foo/bar", "/foo/bar", `/who/are/\*you`, `/\*you`},
		{"/conxxx", "xxx", `/con:tact`, `:tact`},
		{"/conooo/xxx", "ooo", `/con:tact`, `:tact`},
	}

	for _, conflict := range conflicts {
		// I have to re-create a 'tree', because the 'tree' will be
		// in an inconsistent state when the loop recovers from the
		// panic which threw by 'addRoute' function.
		tree := &node{}
		routes := [...]string{
			"/con:tact",
			"/who/are/*you",
			"/who/foo/hello",
		}

		for _, route := range routes {
			tree.addRoute(route, newRoute(route, fakeHandler(route)))
		}

		recv := catchPanic(func() {
			tree.addRoute(conflict.route, newRoute(conflict.route, fakeHandler(conflict.route)))
		})

		reg := regexp.MustCompile(fmt.Sprintf(
			"'%s' in new path .* conflicts with existing wildcard '%s' in existing prefix '%s'",
			conflict.segPath, conflict.existSegPath, conflict.existPath,
		))
		assert.Truef(t, reg.MatchString(fmt.Sprint(recv)), "invalid wildcard conflict error (%v)", recv)
	}
}
