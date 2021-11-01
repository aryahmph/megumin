package service

import (
	"encoding/gob"
	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/Rhymen/go-whatsapp"
	"megumin/pkg/exception"
	"os"
)

type WhatsappService interface {
	ReadSession() (whatsapp.Session, error)
	WriteSession(session whatsapp.Session)
	Login(conn *whatsapp.Conn)
}

type WhatsappServiceImpl struct {
}

func NewWhatsappServiceImpl() *WhatsappServiceImpl {
	return &WhatsappServiceImpl{}
}

func (service *WhatsappServiceImpl) ReadSession() (whatsapp.Session, error) {
	session := whatsapp.Session{}

	file, err := os.Open("sessions/whatsappSession.gob")
	if err != nil {
		return session, err
	}
	defer func(file *os.File) {
		err := file.Close()
		exception.LogIfError(err)
	}(file)

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		return session, err
	}
	exception.LogIfError(err)

	return session, nil
}

func (service *WhatsappServiceImpl) WriteSession(session whatsapp.Session) {
	_, err := os.Stat("sessions")
	if os.IsNotExist(err) {
		err := os.Mkdir("sessions", 0755)
		exception.LogIfError(err)
	}
	exception.LogIfError(err)

	file, err := os.Create("sessions/whatsappSession.gob")
	exception.LogIfError(err)
	defer func(file *os.File) {
		err := file.Close()
		exception.LogIfError(err)
	}(file)

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(session)
	exception.LogIfError(err)
}

func (service *WhatsappServiceImpl) Login(conn *whatsapp.Conn) {
	session, err  := service.ReadSession()
	if err == nil {
		session, err = conn.RestoreWithSession(session)
		exception.LogIfError(err)
	} else {
		qr := make(chan string)
		go func() {
			terminal := qrcodeTerminal.New()
			terminal.Get(<-qr).Print()
		}()

		session, err = conn.Login(qr)
		exception.LogIfError(err)
	}

	service.WriteSession(session)
}
