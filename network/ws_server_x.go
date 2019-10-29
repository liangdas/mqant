package network

import (
	"crypto/tls"
	"github.com/liangdas/mqant/log"
	"golang.org/x/net/websocket"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type WSServer struct {
	Addr        string
	Tls         bool //是否支持tls
	CertFile    string
	KeyFile     string
	MaxConnNum  int
	MaxMsgLen   uint32
	HTTPTimeout time.Duration
	NewAgent    func(*WSConn) Agent
	ln          net.Listener
	handler     *WSHandler
}

type WSHandler struct {
	maxConnNum int
	maxMsgLen  uint32
	newAgent   func(*WSConn) Agent
	mutexConns sync.Mutex
	wg         sync.WaitGroup
}

func (handler *WSHandler) Echo(conn *websocket.Conn) {
	handler.wg.Add(1)
	defer handler.wg.Done()
	conn.PayloadType = websocket.BinaryFrame
	wsConn := newWSConn(conn)
	agent := handler.newAgent(wsConn)
	agent.Run()

	// cleanup
	wsConn.Close()
	handler.mutexConns.Lock()
	handler.mutexConns.Unlock()
	agent.OnClose()
}

func (handler *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	ws := websocket.Server{
		Handler: websocket.Handler(handler.Echo),
		Handshake: func(config *websocket.Config, request *http.Request) error {
			var scheme string
			if request.TLS != nil {
				scheme = "wss"
			} else {
				scheme = "ws"
			}
			config.Origin, _ = url.ParseRequestURI(scheme + "://" + request.RemoteAddr + request.URL.RequestURI())
			config.Protocol = []string{"mqttv3.1"}
			return nil
		},
	}
	ws.ServeHTTP(w, r)
}

func (server *WSServer) Start() {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Warning("%v", err)
	}

	if server.HTTPTimeout <= 0 {
		server.HTTPTimeout = 10 * time.Second
		log.Warning("invalid HTTPTimeout, reset to %v", server.HTTPTimeout)
	}
	if server.NewAgent == nil {
		log.Warning("NewAgent must not be nil")
	}
	if server.Tls {
		tlsConf := new(tls.Config)
		tlsConf.Certificates = make([]tls.Certificate, 1)
		tlsConf.Certificates[0], err = tls.LoadX509KeyPair(server.CertFile, server.KeyFile)
		if err == nil {
			ln = tls.NewListener(ln, tlsConf)
			log.Info("WS Listen TLS load success")
		} else {
			log.Warning("ws_server tls :%v", err)
		}
	}
	server.ln = ln
	server.handler = &WSHandler{
		maxConnNum: server.MaxConnNum,
		maxMsgLen:  server.MaxMsgLen,
		newAgent:   server.NewAgent,
	}
	ws := websocket.Server{
		Handler: websocket.Handler(server.handler.Echo),
		Handshake: func(config *websocket.Config, r *http.Request) error {
			var scheme string
			if r.TLS != nil {
				scheme = "wss"
			} else {
				scheme = "ws"
			}
			config.Origin, _ = url.ParseRequestURI(scheme + "://" + r.RemoteAddr + r.URL.RequestURI())
			offeredProtocol := r.Header.Get("Sec-WebSocket-Protocol")
			config.Protocol = []string{offeredProtocol}
			return nil
		},
	}
	httpServer := &http.Server{
		Addr:           server.Addr,
		Handler:        ws,
		ReadTimeout:    server.HTTPTimeout,
		WriteTimeout:   server.HTTPTimeout,
		MaxHeaderBytes: 1024,
	}
	log.Info("WS Listen :%s", server.Addr)
	go httpServer.Serve(ln)
}

func (server *WSServer) Close() {
	server.ln.Close()

	server.handler.wg.Wait()
}
