Change Log
==========

under development
-----------------
- Addd `Context.GetHeader`, a shortcut of http.Request.Header.Get.

v1.7.0 March 30, 2020
---------------------
- Add `Context.WriteHeader`, an alias of http.ResponseWriter.WriteHeader.
- Add `Context.IsAJAX` to determine whether it is an AJAX request.
- Write error to log.


v1.6.1 March 24, 2020
---------------------
- Bug #21 Fix the call sequence of middleware.
