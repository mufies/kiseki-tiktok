package main

import (
	"io"
	"log"
	"net"

	"github.com/yutopp/go-rtmp"
	rtmpmsg "github.com/yutopp/go-rtmp/message"
)

type TestHandler struct {
	rtmp.DefaultHandler
}

func (h *TestHandler) OnServe(conn *rtmp.Conn) {
	log.Printf("OnServe")
}

func (h *TestHandler) OnConnect(timestamp uint32, cmd *rtmpmsg.NetConnectionConnect) error {
	log.Printf("OnConnect: App=%s", cmd.Command.App)
	return nil
}

func (h *TestHandler) OnCreateStream(timestamp uint32, cmd *rtmpmsg.NetConnectionCreateStream) error {
	log.Printf("OnCreateStream")
	return nil
}

func (h *TestHandler) OnPublish(_ *rtmp.StreamContext, timestamp uint32, cmd *rtmpmsg.NetStreamPublish) error {
	log.Printf("OnPublish: %s", cmd.PublishingName)
	return nil
}

func (h *TestHandler) OnAudio(timestamp uint32, payload io.Reader) error {
	log.Printf("OnAudio")
	return nil
}

func (h *TestHandler) OnVideo(timestamp uint32, payload io.Reader) error {
	log.Printf("OnVideo")
	return nil
}

func (h *TestHandler) OnClose() {
	log.Printf("OnClose")
}

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":1936")
	if err != nil {
		log.Panicf("Failed: %+v", err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Panicf("Failed: %+v", err)
	}

	log.Printf("Test RTMP server listening on :1936")

	srv := rtmp.NewServer(&rtmp.ServerConfig{
		OnConnect: func(conn net.Conn) (io.ReadWriteCloser, *rtmp.ConnConfig) {
			log.Printf("Incoming connection from %s", conn.RemoteAddr())

			h := &TestHandler{}

			return conn, &rtmp.ConnConfig{
				Handler: h,

				ControlState: rtmp.StreamControlStateConfig{
					DefaultBandwidthWindowSize: 6 * 1024 * 1024 / 8,
				},
			}
		},
	})

	if err := srv.Serve(listener); err != nil {
		log.Panicf("Failed: %+v", err)
	}
}
