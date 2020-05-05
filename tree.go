// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package clevergo

import (
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"
)

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func longestCommonPrefix(a, b string) int {
	i := 0
	max := min(len(a), len(b))
	for i < max && a[i] == b[i] {
		i++
	}
	return i
}

// Search for a wildcard segment and check the name for invalid characters.
// Returns -1 as index, if no wildcard was found.
func findWildcard(path string) (wilcard string, i int, valid bool) {
	// Find start
	for start, c := range []byte(path) {
		// A wildcard starts with ':' (param) or '*' (catch-all)
		if c != ':' && c != '*' {
			continue
		}

		// Find end and check for invalid characters
		valid = true
		for end, c := range []byte(path[start+1:]) {
			switch c {
			case '/':
				return path[start : start+1+end], start, valid
			case ':', '*':
				valid = false
			}
		}
		return path[start:], start, valid
	}
	return "", -1, false
}

func countParams(path string) uint16 {
	var n uint
	for i := range []byte(path) {
		switch path[i] {
		case ':', '*':
			n++
		}
	}
	return uint16(n)
}

type nodeType uint8

const (
	static nodeType = iota // default
	root
	param
	catchAll
)

type node struct {
	path      string
	indices   string
	wildChild bool
	nType     nodeType
	priority  uint32
	children  []*node
	route     *Route
}

// Increments priority of the given child and reorders if necessary
func (n *node) incrementChildPrio(pos int) int {
	cs := n.children
	cs[pos].priority++
	prio := cs[pos].priority

	// Adjust position (move to front)
	newPos := pos
	for ; newPos > 0 && cs[newPos-1].priority < prio; newPos-- {
		// Swap node positions
		cs[newPos-1], cs[newPos] = cs[newPos], cs[newPos-1]

	}

	// Build new index char string
	if newPos != pos {
		n.indices = n.indices[:newPos] + // Unchanged prefix, might be empty
			n.indices[pos:pos+1] + // The index char we move
			n.indices[newPos:pos] + n.indices[pos+1:] // Rest without char at 'pos'
	}

	return newPos
}

// addRoute adds a node with the given handle to the path.
// Not concurrency-safe!
func (n *node) addRoute(path string, route *Route) {
	fullPath := path
	n.priority++

	// Empty tree
	if len(n.path) == 0 && len(n.indices) == 0 {
		n.insertChild(path, fullPath, route)
		n.nType = root
		return
	}

walk:
	for {
		// Find the longest common prefix.
		// This also implies that the common prefix contains no ':' or '*'
		// since the existing key can't contain those chars.
		i := longestCommonPrefix(path, n.path)

		// Split edge
		if i < len(n.path) {
			child := node{
				path:      n.path[i:],
				wildChild: n.wildChild,
				nType:     static,
				indices:   n.indices,
				children:  n.children,
				route:     n.route,
				priority:  n.priority - 1,
			}

			n.children = []*node{&child}
			// []byte for proper unicode char conversion, see #65
			n.indices = string([]byte{n.path[i]})
			n.path = path[:i]
			n.route = nil
			n.wildChild = false
		}

		// Make new node a child of this node
		if i < len(path) {
			path = path[i:]

			if n.wildChild {
				n = n.children[0]
				n.priority++

				// Check if the wildcard matches
				if len(path) >= len(n.path) && n.path == path[:len(n.path)] &&
					// Adding a child to a catchAll is not possible
					n.nType != catchAll &&
					// Check for longer wildcard, e.g. :name and :names
					(len(n.path) >= len(path) || path[len(n.path)] == '/') {
					continue walk
				} else {
					// Wildcard conflict
					pathSeg := path
					if n.nType != catchAll {
						pathSeg = strings.SplitN(pathSeg, "/", 2)[0]
					}
					prefix := fullPath[:strings.Index(fullPath, pathSeg)] + n.path
					panic("'" + pathSeg +
						"' in new path '" + fullPath +
						"' conflicts with existing wildcard '" + n.path +
						"' in existing prefix '" + prefix +
						"'")
				}
			}

			idxc := path[0]

			// '/' after param
			if n.nType == param && idxc == '/' && len(n.children) == 1 {
				n = n.children[0]
				n.priority++
				continue walk
			}

			// Check if a child with the next path byte exists
			for i, c := range []byte(n.indices) {
				if c == idxc {
					i = n.incrementChildPrio(i)
					n = n.children[i]
					continue walk
				}
			}

			// Otherwise insert it
			if idxc != ':' && idxc != '*' {
				// []byte for proper unicode char conversion, see #65
				n.indices += string([]byte{idxc})
				child := &node{}
				n.children = append(n.children, child)
				n.incrementChildPrio(len(n.indices) - 1)
				n = child
			}
			n.insertChild(path, fullPath, route)
			return
		}

		// Otherwise add handle to current node
		if n.route != nil {
			panic("a route is already registered for path '" + fullPath + "'")
		}
		n.route = route
		return
	}
}

func (n *node) insertChild(path, fullPath string, route *Route) {
	for {
		// Find prefix until first wildcard
		wildcard, i, valid := findWildcard(path)
		if i < 0 { // No wilcard found
			break
		}

		// The wildcard name must not contain ':' and '*'
		if !valid {
			panic("only one wildcard per path segment is allowed, has: '" +
				wildcard + "' in path '" + fullPath + "'")
		}

		// Check if the wildcard has a name
		if len(wildcard) < 2 {
			panic("wildcards must be named with a non-empty name in path '" + fullPath + "'")
		}

		// Check if this node has existing children which would be
		// unreachable if we insert the wildcard here
		if len(n.children) > 0 {
			panic("wildcard segment '" + wildcard +
				"' conflicts with existing children in path '" + fullPath + "'")
		}

		if wildcard[0] == ':' { // param
			if i > 0 {
				// Insert prefix before the current wildcard
				n.path = path[:i]
				path = path[i:]
			}

			n.wildChild = true
			child := &node{
				nType: param,
				path:  wildcard,
			}
			n.children = []*node{child}
			n = child
			n.priority++

			// If the path doesn't end with the wildcard, then there
			// will be another non-wildcard subpath starting with '/'
			if len(wildcard) < len(path) {
				path = path[len(wildcard):]
				child := &node{
					priority: 1,
				}
				n.children = []*node{child}
				n = child
				continue
			}

			// Otherwise we're done. Insert the handle in the new leaf
			n.route = route
			return

		} else { // catchAll
			if i+len(wildcard) != len(path) {
				panic("catch-all routes are only allowed at the end of the path in path '" + fullPath + "'")
			}

			if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
				panic("catch-all conflicts with existing handle for the path segment root in path '" + fullPath + "'")
			}

			// Currently fixed width 1 for '/'
			i--
			if path[i] != '/' {
				panic("no / before catch-all in path '" + fullPath + "'")
			}

			n.path = path[:i]

			// First node: catchAll node with empty path
			child := &node{
				wildChild: true,
				nType:     catchAll,
			}
			n.children = []*node{child}
			n.indices = string('/')
			n = child
			n.priority++

			// Second node: node holding the variable
			child = &node{
				path:     path[i:],
				nType:    catchAll,
				route:    route,
				priority: 1,
			}
			n.children = []*node{child}

			return
		}
	}

	// If no wildcard was found, simply insert the path and handle
	n.path = path
	n.route = route
}

// Returns the handle registered with the given path (key). The values of
// wildcards are saved to a map.
// If no handle can be found, a TSR (trailing slash redirect) recommendation is
// made if a handle exists with an extra (without the) trailing slash for the
// given path.
func (n *node) getValue(path string, ps *Params, useRawPath bool) (route *Route, tsr bool) {
walk: // Outer loop for walking the tree
	for {
		prefix := n.path
		if len(path) > len(prefix) {
			if path[:len(prefix)] == prefix {
				path = path[len(prefix):]

				// If this node does not have a wildcard (param or catchAll)
				// child, we can just look up the next child node and continue
				// to walk down the tree
				if !n.wildChild {
					idxc := path[0]
					for i, c := range []byte(n.indices) {
						if c == idxc {
							n = n.children[i]
							continue walk
						}
					}

					// Nothing found.
					// We can recommend to redirect to the same URL without a
					// trailing slash if a leaf exists for that path.
					tsr = (path == "/" && n.route != nil)
					return

				}

				// Handle wildcard child
				n = n.children[0]
				switch n.nType {
				case param:
					// Find param end (either '/' or path end)
					end := 0
					for end < len(path) && path[end] != '/' {
						end++
					}

					// Save param value
					if ps != nil {
						// Expand slice within preallocated capacity
						i := len(*ps)
						(*ps) = (*ps)[:i+1]
						param := Param{
							Key: n.path[1:],
						}
						if useRawPath {
							param.Value, _ = url.PathUnescape(path[:end])
						} else {
							param.Value = path[:end]
						}
						(*ps)[i] = param
					}

					// We need to go deeper!
					if end < len(path) {
						if len(n.children) > 0 {
							path = path[end:]
							n = n.children[0]
							continue walk
						}

						// ... but we can't
						tsr = (len(path) == end+1)
						return
					}

					if route = n.route; route != nil {
						return
					} else if len(n.children) == 1 {
						// No handle found. Check if a handle for this path + a
						// trailing slash exists for TSR recommendation
						n = n.children[0]
						tsr = (n.path == "/" && n.route != nil)
					}

					return

				case catchAll:
					// Save param value
					if ps != nil {
						// Expand slice within preallocated capacity
						i := len(*ps)
						(*ps) = (*ps)[:i+1]
						param := Param{
							Key: n.path[2:],
						}
						if useRawPath {
							param.Value, _ = url.PathUnescape(path)
						} else {
							param.Value = path
						}
						(*ps)[i] = param
					}

					route = n.route
					return

				default:
					panic("invalid node type")
				}
			}
		} else if path == prefix {
			// We should have reached the node containing the handle.
			// Check if this node has a handle registered.
			if route = n.route; route != nil {
				return
			}

			// If there is no handle for this route, but this route has a
			// wildcard child, there must be a handle for this path with an
			// additional trailing slash
			if path == "/" && n.wildChild && n.nType != root {
				tsr = true
				return
			}

			// No handle found. Check if a handle for this path + a
			// trailing slash exists for trailing slash recommendation
			for i, c := range []byte(n.indices) {
				if c == '/' {
					n = n.children[i]
					tsr = (len(n.path) == 1 && n.route != nil) ||
						(n.nType == catchAll && n.children[0].route != nil)
					return
				}
			}
			return
		}

		// Nothing found. We can recommend to redirect to the same URL with an
		// extra trailing slash if a leaf exists for that path
		tsr = (path == "/") ||
			(len(prefix) == len(path)+1 && prefix[len(path)] == '/' &&
				path == prefix[:len(prefix)-1] && n.route != nil)
		return
	}
}

// Makes a case-insensitive lookup of the given path and tries to find a handler.
// It can optionally also fix trailing slashes.
// It returns the case-corrected path and a bool indicating whether the lookup
// was successful.
func (n *node) findCaseInsensitivePath(path string, fixTrailingSlash bool) (fixedPath string, found bool) {
	const stackBufSize = 128

	// Use a static sized buffer on the stack in the common case.
	// If the path is too long, allocate a buffer on the heap instead.
	buf := make([]byte, 0, stackBufSize)
	if l := len(path) + 1; l > stackBufSize {
		buf = make([]byte, 0, l)
	}

	ciPath := n.findCaseInsensitivePathRec(
		path,
		buf,       // Preallocate enough memory for new path
		[4]byte{}, // Empty rune buffer
		fixTrailingSlash,
	)

	return string(ciPath), ciPath != nil
}

// Shift bytes in array by n bytes left
func shiftNRuneBytes(rb [4]byte, n int) [4]byte {
	switch n {
	case 0:
		return rb
	case 1:
		return [4]byte{rb[1], rb[2], rb[3], 0}
	case 2:
		return [4]byte{rb[2], rb[3]}
	case 3:
		return [4]byte{rb[3]}
	default:
		return [4]byte{}
	}
}

// Recursive case-insensitive lookup function used by n.findCaseInsensitivePath
func (n *node) findCaseInsensitivePathRec(path string, ciPath []byte, rb [4]byte, fixTrailingSlash bool) []byte {
	npLen := len(n.path)

walk: // Outer loop for walking the tree
	for len(path) >= npLen && (npLen == 0 || strings.EqualFold(path[1:npLen], n.path[1:])) {
		// Add common prefix to result
		oldPath := path
		path = path[npLen:]
		ciPath = append(ciPath, n.path...)

		if len(path) > 0 {
			// If this node does not have a wildcard (param or catchAll) child,
			// we can just look up the next child node and continue to walk down
			// the tree
			if !n.wildChild {
				// Skip rune bytes already processed
				rb = shiftNRuneBytes(rb, npLen)

				if rb[0] != 0 {
					// Old rune not finished
					idxc := rb[0]
					for i, c := range []byte(n.indices) {
						if c == idxc {
							// continue with child node
							n = n.children[i]
							npLen = len(n.path)
							continue walk
						}
					}
				} else {
					// Process a new rune
					var rv rune

					// Find rune start.
					// Runes are up to 4 byte long,
					// -4 would definitely be another rune.
					var off int
					for max := min(npLen, 3); off < max; off++ {
						if i := npLen - off; utf8.RuneStart(oldPath[i]) {
							// read rune from cached path
							rv, _ = utf8.DecodeRuneInString(oldPath[i:])
							break
						}
					}

					// Calculate lowercase bytes of current rune
					lo := unicode.ToLower(rv)
					utf8.EncodeRune(rb[:], lo)

					// Skip already processed bytes
					rb = shiftNRuneBytes(rb, off)

					idxc := rb[0]
					for i, c := range []byte(n.indices) {
						// Lowercase matches
						if c == idxc {
							// must use a recursive approach since both the
							// uppercase byte and the lowercase byte might exist
							// as an index
							if out := n.children[i].findCaseInsensitivePathRec(
								path, ciPath, rb, fixTrailingSlash,
							); out != nil {
								return out
							}
							break
						}
					}

					// If we found no match, the same for the uppercase rune,
					// if it differs
					if up := unicode.ToUpper(rv); up != lo {
						utf8.EncodeRune(rb[:], up)
						rb = shiftNRuneBytes(rb, off)

						idxc := rb[0]
						for i, c := range []byte(n.indices) {
							// Uppercase matches
							if c == idxc {
								// Continue with child node
								n = n.children[i]
								npLen = len(n.path)
								continue walk
							}
						}
					}
				}

				// Nothing found. We can recommend to redirect to the same URL
				// without a trailing slash if a leaf exists for that path
				if fixTrailingSlash && path == "/" && n.route != nil {
					return ciPath
				}
				return nil
			}

			n = n.children[0]
			switch n.nType {
			case param:
				// Find param end (either '/' or path end)
				end := 0
				for end < len(path) && path[end] != '/' {
					end++
				}

				// Add param value to case insensitive path
				ciPath = append(ciPath, path[:end]...)

				// We need to go deeper!
				if end < len(path) {
					if len(n.children) > 0 {
						// Continue with child node
						n = n.children[0]
						npLen = len(n.path)
						path = path[end:]
						continue
					}

					// ... but we can't
					if fixTrailingSlash && len(path) == end+1 {
						return ciPath
					}
					return nil
				}

				if n.route != nil {
					return ciPath
				} else if fixTrailingSlash && len(n.children) == 1 {
					// No handle found. Check if a handle for this path + a
					// trailing slash exists
					n = n.children[0]
					if n.path == "/" && n.route != nil {
						return append(ciPath, '/')
					}
				}
				return nil

			case catchAll:
				return append(ciPath, path...)

			default:
				panic("invalid node type")
			}
		} else {
			// We should have reached the node containing the handle.
			// Check if this node has a handle registered.
			if n.route != nil {
				return ciPath
			}

			// No handle found.
			// Try to fix the path by adding a trailing slash
			if fixTrailingSlash {
				for i, c := range []byte(n.indices) {
					if c == '/' {
						n = n.children[i]
						if (len(n.path) == 1 && n.route != nil) ||
							(n.nType == catchAll && n.children[0].route != nil) {
							return append(ciPath, '/')
						}
						return nil
					}
				}
			}
			return nil
		}
	}

	// Nothing found.
	// Try to fix the path by adding / removing a trailing slash
	if fixTrailingSlash {
		if path == "/" {
			return ciPath
		}
		if len(path)+1 == npLen && n.path[len(path)] == '/' &&
			strings.EqualFold(path[1:], n.path[1:len(path)]) && n.route != nil {
			return append(ciPath, n.path...)
		}
	}
	return nil
}
