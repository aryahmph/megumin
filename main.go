package main

import (
	"github.com/Rhymen/go-whatsapp"
	"megumin/handler"
	"megumin/pkg/exception"
	"megumin/service"
	"os"
	"os/signal"
	"regexp"
	"sync"
	"syscall"
	"time"
)

func main() {
	conn, err := whatsapp.NewConn(5 * time.Second)
	exception.LogIfError(err)
	conn.SetClientVersion(3, 3234, 9)

	now := time.Now().Unix()
	regexGroupId, err := regexp.Compile(`@g.us$`)
	regexNumber, err := regexp.Compile(`^[0-9]*$`)
	exception.LogIfError(err)

	rwMutex := new(sync.RWMutex)
	gameService := service.NewGameServiceImpl(rwMutex)

	whatsappService := service.NewWhatsappServiceImpl()
	whatsappHandler := handler.NewWhatsappHandlerImpl(conn, uint64(now), regexGroupId, regexNumber, gameService)

	whatsappService.Login(conn)
	<-time.After(3 * time.Second)

	conn.AddHandler(whatsappHandler)

	errChan := make(chan os.Signal, 1)
	signal.Notify(errChan, os.Interrupt, syscall.SIGTERM)
	<-errChan
}
