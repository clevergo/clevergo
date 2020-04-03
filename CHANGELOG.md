Change Log
==========

under development
-----------------
- Add `Context.JSON` to send JSON response.
- Add `Context.String` to send string response.
- Add `Context.XML` to send XML response.
- `Content.SetContentTypeJSON` and `Content.SetContentTypeXML` append `charset=utf-8` to content type header.
- Add `Context.HTML` to send HTML response.
- Add `Context.Cookie` and `Context.Cookies`.
- Add `Context.FormValue`.
- Add `Context.PostFormValue`.
- Add `Context.QueryString`.

v1.8.1 April 2, 2020
--------------------
- Fix `WrapHH` doesn't returns the error of final handle.

v1.8.0 April 2, 2020
--------------------
- Add `Context.GetHeader`, a shortcut of http.Request.Header.Get.
- Add `WrapH` to wrap a HTTP handler as a middleware.
- Add `WrapHH` to wrap func(http.Handler) http.Handler as a middleware.

v1.7.0 March 30, 2020
---------------------
- Add `Context.WriteHeader`, an alias of http.ResponseWriter.WriteHeader.
- Add `Context.IsAJAX` to determine whether it is an AJAX request.
- Write error to log.

v1.6.1 March 24, 2020
---------------------
- Bug #21 Fix the call sequence of middleware.
