package main

import (
	"flag"
	"go-one/common/context"
	"go-one/common/log"
	"go-one/demo/game/processor"
	"go-one/game"
	"go-one/game/player"
	"go-one/game/processor_center"
	"golang.org/x/net/websocket"
	"net/http"
	_ "net/http/pprof"
)

func main() {

	flag.Parse()

	context.Init()

	game.InitConfig()

	RegisterProcessor()

	gameServer := game.NewGameServer()

	player.SetPlayerType(&DemoPlayer{})

	go setupHTTPServer(":8833", nil, "", "")

	gameServer.Run()

}

func RegisterProcessor() {
	processor_center.GPM.RegisterProcessor(&processor.Test{})
}

type DemoPlayer struct {
	player.Player
}

func (p *DemoPlayer) OnCreated() {

}
func (p *DemoPlayer) OnDestroy() {

}

func (p *DemoPlayer) OnClientConnected() {

}
func (p *DemoPlayer) OnClientDisconnected() {

}

func setupHTTPServer(listenAddr string, wsHandler func(ws *websocket.Conn), certFile string, keyFile string) {
	log.Infof("http server listening on %s", listenAddr)
	log.Infof("pprof http://%s/debug/pprof/ ... available commands: ", listenAddr)
	log.Infof("    go tool pprof http://%s/debug/pprof/heap", listenAddr)
	log.Infof("    go tool pprof http://%s/debug/pprof/profile", listenAddr)
	if keyFile != "" || certFile != "" {
		log.Infof("TLS is enabled on http: key=%s, cert=%s", keyFile, certFile)
	}

	if wsHandler != nil {
		http.Handle("/ws", websocket.Handler(wsHandler))
	}

	go func() {
		if keyFile == "" && certFile == "" {
			http.ListenAndServe(listenAddr, nil)
		} else {
			http.ListenAndServeTLS(listenAddr, certFile, keyFile, nil)
		}
	}()
}
