Change Log
==========

under development
-----------------
- Add `Context.JSON` to sends JSON response.
- Add `Context.String` to sends string response.
- Add `Context.XML` to sends XML response.

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
