// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gem

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-gem/log"
	"github.com/go-gem/sessions"
	"github.com/valyala/fasthttp"
)

const (
	// Gem name
	name = "Gem"

	// Gem version
	version = "v1.0.0-alpha"
)

// Name returns name.
func Name() string {
	return name
}

// Version returns current version of Gem.
func Version() string {
	return version
}

// serverName returns default server name.
func serverName() string {
	return name + "/" + version
}

var (
	mutex sync.RWMutex

	serversWg *sync.WaitGroup

	servers map[string]*Server

	isGracefulRestart bool

	isForked bool

	isShuttingDown bool

	serversAddr []string

	// serversFdOffset for store server listener's file descriptor
	serversFdOffset map[string]uint
)

func init() {
	mutex = sync.RWMutex{}

	serversWg = &sync.WaitGroup{}

	servers = make(map[string]*Server)

	isGracefulRestart = os.Getenv("GEM_GRACEFUL_RESTART") == "true"

	serversAddr = make([]string, 0)
	serversFdOffset = make(map[string]uint)

	initServersFdOffset()
}

func initServersFdOffset() {
	if addrs := os.Getenv("GEM_SERVER_ADDRS"); len(addrs) > 0 {
		serversAddr = strings.Split(addrs, ",")
		for i, addr := range serversAddr {
			serversFdOffset[addr] = uint(i)
		}
	}
}

var (
	waitTimeout = time.Second * 15

	waitTimeoutError = errors.New("timeout")

	logger = log.New(os.Stderr, log.Llongfile|log.LstdFlags, log.LevelAll)
)

// New returns a Server instance with default setting.
func New(addr string, handler HandlerFunc) *Server {
	mutex.Lock()
	defer mutex.Unlock()

	serversWg.Add(1)

	if addr == "" {
		addr = ":http"
	}

	srv := &Server{
		addr: addr,
		server: &fasthttp.Server{
			Name: serverName(),
		},
		wg:      &sync.WaitGroup{},
		sigChan: make(chan os.Signal),
		signals: map[os.Signal]Signal{
			syscall.SIGHUP:  SignalRestart,
			syscall.SIGTERM: SignalShutdown,
		},
		logger:      logger,
		waitTimeout: waitTimeout,
	}

	// Initialize handler.
	srv.init(handler)

	servers[addr] = srv

	return srv
}

// Signal
type Signal int8

// Signals
const (
	_              = iota
	SignalShutdown = iota
	SignalRestart
)

func isSignal(sig Signal) bool {
	return sig == SignalShutdown || sig == SignalRestart
}

// Server
type Server struct {
	server        *fasthttp.Server
	addr          string
	listener      net.Listener
	wg            *sync.WaitGroup
	sigChan       chan os.Signal
	signals       map[os.Signal]Signal
	waitTimeout   time.Duration
	logger        Logger
	sessionsStore sessions.Store
}

func (srv *Server) SetSignal(sig1 os.Signal, sig2 Signal) error {
	if !isSignal(sig2) {
		return fmt.Errorf("unsupported signal: %v", sig2)
	}

	srv.signals[sig1] = sig2
	return nil
}

// SetLogger set logger.
func (srv *Server) SetLogger(logger Logger) {
	srv.logger = logger
}

// SetSessionsStore set sessions store.
func (srv *Server) SetSessionsStore(store sessions.Store) {
	srv.sessionsStore = store
}

// SetWaitTimeout wait timeout.
func (srv *Server) SetWaitTimeout(duration time.Duration) {
	srv.waitTimeout = duration
}

// getFileListener get file listener by the given addr.
func getFileListener(addr string) (net.Listener, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	offset := serversFdOffset[addr]
	f := os.NewFile(uintptr(3+offset), "")
	return net.FileListener(f)
}

type listenerConfig struct {
	net      string
	mode     os.FileMode
	certData []byte
	keyData  []byte
	certFile string
	keyFile  string
}

// initListener initialize listener.
func (srv *Server) initListener(config listenerConfig) (err error) {
	if isGracefulRestart {
		srv.listener, err = getFileListener(srv.addr)
		if err != nil {
			return err
		}

		// Kill parent process.
		if len(servers) == len(serversAddr) {
			syscall.Kill(os.Getppid(), syscall.SIGTERM)
		}

		return
	}

	switch config.net {
	case "tcp4":
		srv.listener, err = net.Listen("tcp4", srv.addr)
	case "unix":
		if err = os.Remove(srv.addr); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("unexpected error when trying to remove unix socket file %q: %s", srv.addr, err)
		}
		srv.listener, err = net.Listen("unix", srv.addr)
		if err != nil {
			return err
		}
		if err = os.Chmod(srv.addr, config.mode); err != nil {
			return fmt.Errorf("cannot chmod %#o for %q: %s", config.mode, srv.addr, err)
		}
	case "tls":
		srv.listener, err = net.Listen("tcp4", srv.addr)
		if err != nil {
			return err
		}
		srv.listener, err = newTLSListener(srv.listener, config.certFile, config.keyFile)
		if err != nil {
			return err
		}
	default:
		err = fmt.Errorf("Unsupported server address: %s\n", srv.addr)
	}

	return
}

// ListenAndServe serves HTTP requests from the given TCP4 addr.
func (srv *Server) ListenAndServe() error {
	go srv.handleSignals()

	if err := srv.initListener(listenerConfig{net: "tcp4"}); err != nil {
		return err
	}

	return srv.serve()
}

// handleSignals handle signals.
func (srv *Server) handleSignals() {
	var sig os.Signal
	for sig, _ = range srv.signals {
		signal.Notify(srv.sigChan, sig)
	}

	pid := syscall.Getpid()
	for {
		sig = <-srv.sigChan
		switch srv.signals[sig] {
		case SignalRestart:
			err := fork()
			if err != nil {
				log.Printf("[%d] Fork err: %s", pid, err)
			}
			srv.stop()
		case SignalShutdown:
			if err := srv.wait(); err != nil {
				log.Printf(
					"[%d] Server(%s) has been shutdown, but some exsiting connctions reach error: %s.\n",
					pid, srv.addr, err,
				)
			} else {
				log.Printf("[%d] Server(%s) shutdown successfully.\n", pid, srv.addr)
			}
			serversWg.Done()
			shutdown()
			return
		default:
			log.Printf("[%d] Received %v.\n", pid, sig)
		}
	}
}

func (srv *Server) serve() error {
	for {
		conn, err := srv.listener.Accept()
		if err != nil {
			return err
		}
		go srv.ServeConn(conn)
	}
}

func (srv *Server) ServeConn(conn net.Conn) (err error) {
	srv.wg.Add(1)
	defer func() {
		conn.Close()
		srv.wg.Done()
	}()

	err = srv.server.ServeConn(conn)
	if err != nil {
		srv.logger.Errorf("Serve conn error: %s\n", err)
	}

	return
}

// init initialize server.
func (srv *Server) init(handler HandlerFunc) {
	srv.server.Handler = func(reqCtx *fasthttp.RequestCtx) {
		ctx := acquireContext(srv, reqCtx)
		defer releaseContext(ctx)
		handler(ctx)
	}
}

// Stop stop accepting any incoming connections.
func (srv *Server) stop() {
	// Disable keep-alive of existing connections.
	srv.server.DisableKeepalive = true
}

// wait wait a duration for existing connections to finish,
// returns waitTimeoutError when timeout.
func (srv *Server) wait() error {
	timeout := time.NewTimer(srv.waitTimeout)
	wait := make(chan struct{})
	go func() {
		srv.wg.Wait()
		wait <- struct{}{}
	}()

	select {
	case <-timeout.C:
		return waitTimeoutError
	case <-wait:
		return nil
	}
}

// shutdown shutdown all servers.
func shutdown() {
	mutex.Lock()
	if isShuttingDown {
		return
	}
	isShuttingDown = true
	mutex.Unlock()

	serversWg.Wait()

	log.Printf("[%d] All of old servers have been shutdown successfully.\n", os.Getpid())

	os.Exit(0)
}

func fork() (err error) {
	mutex.Lock()
	if isForked {
		mutex.Unlock()
		return
	}
	isForked = true
	mutex.Unlock()

	pid := syscall.Getpid()
	log.Printf("[%d] Forking...\n", pid)

	files := make([]uintptr, len(servers)+3)
	files = append(files, os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd())
	var addrs = make([]string, 0)
	for _, srv := range servers {
		var f *os.File
		switch srv.listener.(type) {
		case *net.TCPListener:
			f, _ = srv.listener.(*net.TCPListener).File()
			files = append(files)
		case *net.UnixListener:
			f, _ = srv.listener.(*net.UnixListener).File()
		}
		files[len(addrs)+3] = f.Fd()
		addrs = append(addrs, srv.addr)
	}

	env := append(
		os.Environ(),
		"GEM_GRACEFUL_RESTART=true",
	)
	if len(servers) > 1 {
		env = append(env, fmt.Sprintf(`GEM_SERVER_ADDRS=%s`, strings.Join(addrs, ",")))
	}

	execSpec := &syscall.ProcAttr{
		Env:   env,
		Files: files,
	}
	// Fork exec the new version of your server
	fork, err := syscall.ForkExec(os.Args[0], os.Args, execSpec)
	if err != nil {
		return err
	}
	log.Printf("[%d] Fork-exec to %d.\n", pid, fork)

	return
}

// ListenAndServe serves HTTP requests from the given TCP addr
// using the given handler.
func ListenAndServe(addr string, handler HandlerFunc) error {
	srv := New(addr, handler)
	return srv.ListenAndServe()
}

// ListenAndServeUNIX serves HTTP requests from the given UNIX addr.
//
// The function deletes existing file at addr before starting serving.
//
// The server sets the given file mode for the UNIX addr.
func (srv *Server) ListenAndServeUNIX(mode os.FileMode) error {
	go srv.handleSignals()

	if err := srv.initListener(listenerConfig{net: "unix", mode: mode}); err != nil {
		return err
	}

	return srv.serve()
}

// ListenAndServeUNIX serves HTTP requests from the given UNIX addr
// using the given handler.
//
// The function deletes existing file at addr before starting serving.
//
// The server sets the given file mode for the UNIX addr.
func ListenAndServeUNIX(addr string, mode os.FileMode, handler HandlerFunc) error {
	srv := New(addr, handler)
	return srv.ListenAndServeUNIX(mode)
}

// ListenAndServeTLS serves HTTPS requests from the given TCP4 addr.
//
// certFile and keyFile are paths to TLS certificate and key files.
//
// Pass custom listener to Serve if you need listening on non-TCP4 media
// such as IPv6.
func (srv *Server) ListenAndServeTLS(certFile, keyFile string) error {
	go srv.handleSignals()

	if err := srv.initListener(listenerConfig{net: "tls", certFile: certFile, keyFile: keyFile}); err != nil {
		return err
	}

	return srv.serve()
}

// ListenAndServeTLS serves HTTPS requests from the given TCP addr
// using the given handler.
//
// certFile and keyFile are paths to TLS certificate and key files.
func ListenAndServeTLS(addr, certFile, keyFile string, handler HandlerFunc) error {
	srv := New(addr, handler)
	return srv.ListenAndServeTLS(certFile, keyFile)
}

func newTLSListener(ln net.Listener, certFile, keyFile string) (net.Listener, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("cannot load TLS key pair from certFile=%q and keyFile=%q: %s", certFile, keyFile, err)
	}
	return newCertListener(ln, &cert), nil
}

func newCertListener(ln net.Listener, cert *tls.Certificate) net.Listener {
	tlsConfig := &tls.Config{
		Certificates:             []tls.Certificate{*cert},
		PreferServerCipherSuites: true,
	}
	return tls.NewListener(ln, tlsConfig)
}

// ServerConfig see fasthttp.Server for more details.
type ServerConfig struct {
	// Server name for sending in response headers.
	//
	// Default server name is used if left blank.
	Name string `json:"name"`

	// The maximum number of concurrent connections the server may serve.
	//
	// DefaultConcurrency is used if not set.
	Concurrency int `json:"concurrency"`

	// Whether to disable keep-alive connections.
	//
	// The server will close all the incoming connections after sending
	// the first response to client if this option is set to true.
	//
	// By default keep-alive connections are enabled.
	DisableKeepalive bool `json:"disable_keepalive"`

	// Per-connection buffer size for requests' reading.
	// This also limits the maximum header size.
	//
	// Increase this buffer if your clients send multi-KB RequestURIs
	// and/or multi-KB headers (for example, BIG cookies).
	//
	// Default buffer size is used if not set.
	ReadBufferSize int `json:"read_buffer_size"`

	// Per-connection buffer size for responses' writing.
	//
	// Default buffer size is used if not set.
	WriteBufferSize int `json:"write_buffer_size"`

	// Maximum duration for reading the full request (including body).
	//
	// This also limits the maximum duration for idle keep-alive
	// connections.
	//
	// By default request read timeout is unlimited.
	ReadTimeout time.Duration `json:"read_timeout"`

	// Maximum duration for writing the full response (including body).
	//
	// By default response write timeout is unlimited.
	WriteTimeout time.Duration `json:"write_timeout"`

	// Maximum number of concurrent client connections allowed per IP.
	//
	// By default unlimited number of concurrent connections
	// may be established to the server from a single IP address.
	MaxConnsPerIP int `json:"max_conns_per_ip"`

	// Maximum number of requests served per connection.
	//
	// The server closes connection after the last request.
	// 'Connection: close' header is added to the last response.
	//
	// By default unlimited number of requests may be served per connection.
	MaxRequestsPerConn int `json:"max_requests_per_conn"`

	// Maximum keep-alive connection lifetime.
	//
	// The server closes keep-alive connection after its' lifetime
	// expiration.
	//
	// See also ReadTimeout for limiting the duration of idle keep-alive
	// connections.
	//
	// By default keep-alive connection lifetime is unlimited.
	MaxKeepaliveDuration time.Duration `json:"max_keepalive_duration"`

	// Maximum request body size.
	//
	// The server rejects requests with bodies exceeding this limit.
	//
	// Request body size is limited by DefaultMaxRequestBodySize by default.
	MaxRequestBodySize int `json:"max_request_body_size"`

	// Aggressively reduces memory usage at the cost of higher CPU usage
	// if set to true.
	//
	// Try enabling this option only if the server consumes too much memory
	// serving mostly idle keep-alive connections. This may reduce memory
	// usage by more than 50%.
	//
	// Aggressive memory usage reduction is disabled by default.
	ReduceMemoryUsage bool `json:"reduce_memory_usage"`

	// Rejects all non-GET requests if set to true.
	//
	// This option is useful as anti-DoS protection for servers
	// accepting only GET requests. The request size is limited
	// by ReadBufferSize if GetOnly is set.
	//
	// Server accepts all the requests by default.
	GetOnly bool `json:"get_only"`

	// Header names are passed as-is without normalization
	// if this option is set.
	//
	// Disabled header names' normalization may be useful only for proxying
	// incoming requests to other servers expecting case-sensitive
	// header names. See https://github.com/valyala/fasthttp/issues/57
	// for details.
	//
	// By default request and response header names are normalized, i.e.
	// The first letter and the first letters following dashes
	// are uppercased, while all the other letters are lowercased.
	// Examples:
	//
	//     * HOST -> Host
	//     * content-type -> Content-Type
	//     * cONTENT-lenGTH -> Content-Length
	DisableHeaderNamesNormalizing bool `json:"disable_header_names_normalizing"`
}

// LoadConfig load server configuration.
func (srv *Server) LoadConfig(config *ServerConfig) {
	if config.Name != "" {
		srv.server.Name = config.Name
	}
	if config.Concurrency > 0 {
		srv.server.Concurrency = config.Concurrency
	}
	if config.ReadBufferSize > 0 {
		srv.server.ReadBufferSize = config.ReadBufferSize
	}
	if config.WriteBufferSize > 0 {
		srv.server.WriteBufferSize = config.WriteBufferSize
	}
	if config.ReadTimeout > 0 {
		srv.server.ReadTimeout = config.ReadTimeout
	}
	if config.WriteTimeout > 0 {
		srv.server.WriteTimeout = config.WriteTimeout
	}
	if config.MaxConnsPerIP > 0 {
		srv.server.MaxConnsPerIP = config.MaxConnsPerIP
	}
	if config.MaxRequestsPerConn > 0 {
		srv.server.MaxRequestsPerConn = config.MaxRequestsPerConn
	}
	if config.MaxKeepaliveDuration > 0 {
		srv.server.MaxKeepaliveDuration = config.MaxKeepaliveDuration
	}
	if config.MaxRequestBodySize > 0 {
		srv.server.MaxRequestBodySize = config.MaxRequestBodySize
	}
	srv.server.DisableKeepalive = config.DisableKeepalive
	srv.server.ReduceMemoryUsage = config.ReduceMemoryUsage
	srv.server.GetOnly = config.GetOnly
	srv.server.DisableHeaderNamesNormalizing = config.DisableHeaderNamesNormalizing
}
