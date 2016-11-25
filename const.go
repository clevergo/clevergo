// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gem

// Constants
const (
	// AcceptEncoding
	HeaderAcceptEncoding        = "Accept-Encoding"
	HeaderAcceptEncodingDeflate = "deflate"
	HeaderAcceptEncodingGzip    = "gzip"

	// Cross-Origin Resource Sharing Headers
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"

	HeaderAuthorization = "Authorization"

	HeaderBasic = "Basic"

	HeaderBearer = "Bearer"

	HeaderContentEncoding = "Content-Encoding"

	HeaderContentLength = "Content-Length"

	// Content-Types
	HeaderContentType      = "Content-Type"
	HeaderContentTypeForm  = "application/x-www-form-urlencoded"
	HeaderContentTypeHTML  = "text/html; charset=utf-8"
	HeaderContentTypeJSON  = "application/json; charset=utf-8"
	HeaderContentTypeJSONP = "application/javascript; charset=utf-8"
	HeaderContentTypeText  = "text/plain; charset=utf-8"
	HeaderContentTypeXML   = "application/xml; charset=utf-8"

	HeaderOrigin = "Origin"

	HeaderVary = "Vary"

	HeaderXMLHttpRequest = "XMLHttpRequest"
	HeaderXRequestedWith = "X-Requested-With"

	// Methods
	MethodConnect = "CONNECT"
	MethodDelete  = "DELETE"
	MethodGet     = "GET"
	MethodHead    = "HEAD"
	MethodOptions = "OPTIONS"
	MethodPatch   = "PATCH"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodTrace   = "TRACE"
)

// Bytes
var (
	// AcceptEncoding
	HeaderAcceptEncodingBytes        = []byte(HeaderAcceptEncoding)
	HeaderAcceptEncodingDeflateBytes = []byte(HeaderAcceptEncodingDeflate)
	HeaderAcceptEncodingGzipBytes    = []byte(HeaderAcceptEncodingGzip)

	// Cross-Origin Resource Sharing Headers
	HeaderAccessControlAllowCredentialsBytes = []byte(HeaderAccessControlAllowCredentials)
	HeaderAccessControlAllowHeadersBytes     = []byte(HeaderAccessControlAllowHeaders)
	HeaderAccessControlAllowMethodsBytes     = []byte(HeaderAccessControlAllowMethods)
	HeaderAccessControlAllowOriginBytes      = []byte(HeaderAccessControlAllowOrigin)
	HeaderAccessControlExposeHeadersBytes    = []byte(HeaderAccessControlExposeHeaders)
	HeaderAccessControlMaxAgeBytes           = []byte(HeaderAccessControlMaxAge)
	HeaderAccessControlRequestHeadersBytes   = []byte(HeaderAccessControlRequestHeaders)
	HeaderAccessControlRequestMethodBytes    = []byte(HeaderAccessControlRequestMethod)

	HeaderAuthorizationBytes = []byte(HeaderAuthorization)

	HeaderBasicBytes = []byte(HeaderBasic)

	HeaderBearerBytes = []byte(HeaderBearer)

	HeaderContentEncodingBytes = []byte(HeaderContentEncoding)

	HeaderContentLengthBytes = []byte(HeaderContentLength)

	// Content-Types
	HeaderContentTypeBytes      = []byte(HeaderContentType)
	HeaderContentTypeHTMLBytes  = []byte(HeaderContentTypeHTML)
	HeaderContentTypeJSONBytes  = []byte(HeaderContentTypeJSON)
	HeaderContentTypeJSONPBytes = []byte(HeaderContentTypeJSONP)
	HeaderContentTypeTextBytes  = []byte(HeaderContentTypeText)
	HeaderContentTypeXMLBytes   = []byte(HeaderContentTypeXML)

	HeaderOriginBytes = []byte(HeaderOrigin)

	HeaderVaryBytes = []byte(HeaderVary)

	HeaderXMLHttpRequestBytes = []byte(HeaderXMLHttpRequest)
	HeaderXRequestedWithBytes = []byte(HeaderXRequestedWith)

	// Methods
	MethodConnectBytes = []byte(MethodConnect)
	MethodDeleteBytes  = []byte(MethodDelete)
	MethodGetBytes     = []byte(MethodGet)
	MethodHeadBytes    = []byte(MethodHead)
	MethodOptionsBytes = []byte(MethodOptions)
	MethodPatchBytes   = []byte(MethodPatch)
	MethodPostBytes    = []byte(MethodPost)
	MethodPutBytes     = []byte(MethodPut)
	MethodTraceBytes   = []byte(MethodTrace)
)

// Byte units.
const (
	B = 1 << (10 * iota)
	KB
	MB
	GB
	TB
	PB
	EB
)
