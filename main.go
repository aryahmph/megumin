package main

import (
	"github.com/Rhymen/go-whatsapp"
	"megumin/handler"
	"megumin/pkg/exception"
	"megumin/service"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	conn, err := whatsapp.NewConn(5 * time.Second)
	exception.LogIfError(err)
	conn.SetClientVersion(2, 2121, 17)

	whatsappService := service.NewWhatsappServiceImpl()
	whatsappHandler := handler.NewWhatsappHandlerImpl(conn)
	conn.AddHandler(whatsappHandler)

	whatsappService.Login(conn)
	<-time.After(3 * time.Second)

	errChan := make(chan os.Signal, 1)
	signal.Notify(errChan, os.Interrupt, syscall.SIGTERM)
	<-errChan
}
