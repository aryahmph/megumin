package main

import (
	"github.com/Rhymen/go-whatsapp"
	"megumin/handler"
	"megumin/pkg/exception"
	"megumin/service"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"
)

func main() {
	conn, err := whatsapp.NewConn(5 * time.Second)
	exception.LogIfError(err)
	conn.SetClientVersion(3, 3234, 9)

	now := time.Now().Unix()
	regexGroupId, err := regexp.Compile(`@g.us$`)

	whatsappService := service.NewWhatsappServiceImpl()
	whatsappHandler := handler.NewWhatsappHandlerImpl(conn, uint64(now), regexGroupId)

	whatsappService.Login(conn)
	<-time.After(3 * time.Second)

	conn.AddHandler(whatsappHandler)

	errChan := make(chan os.Signal, 1)
	signal.Notify(errChan, os.Interrupt, syscall.SIGTERM)
	<-errChan
}
