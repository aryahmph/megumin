package handler

import (
	"encoding/hex"
	"encoding/json"
	"github.com/Rhymen/go-whatsapp"
	"github.com/Rhymen/go-whatsapp/binary/proto"
	"log"
	"math/rand"
	"megumin/entity"
	"megumin/pkg/exception"
	"regexp"
	"strings"
	"time"
)

type WhatsappHandler interface {
	HandleError(err error)
	HandleTextMessage(message whatsapp.TextMessage)
	generateWebMessageInfo(message whatsapp.TextMessage, text string, participantsJIDs []string) *proto.WebMessageInfo
	getGroupParticipantsJIDs(remoteJid string) []string
}

type WhatsappHandlerImpl struct {
	conn                 *whatsapp.Conn
	lastMessageTimestamp uint64
	regexGroupId         *regexp.Regexp
}

func NewWhatsappHandlerImpl(conn *whatsapp.Conn, lastMessageTimestamp uint64, regexGroupId *regexp.Regexp) *WhatsappHandlerImpl {
	return &WhatsappHandlerImpl{conn: conn, lastMessageTimestamp: lastMessageTimestamp, regexGroupId: regexGroupId}
}

func (handler *WhatsappHandlerImpl) HandleError(err error) {
	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok {
		log.Printf("Connection failed, underlying error: %v", e.Err)
		log.Println("Waiting 30sec...")

		<-time.After(30 * time.Second)

		log.Println("Reconnecting...")
		err := handler.conn.Restore()
		exception.LogIfError(err)
	} else {
		log.Printf("error occoured: %v\n", err)
	}
}

func (handler *WhatsappHandlerImpl) HandleTextMessage(message whatsapp.TextMessage) {
	if message.Info.FromMe || message.Info.Timestamp < handler.lastMessageTimestamp {
		return
	}

	if message.Text == "!absen" && handler.regexGroupId.MatchString(message.Info.RemoteJid) {
		_, err := handler.conn.Read(message.Info.RemoteJid, message.Info.Id)
		exception.LogIfError(err)

		text := "[📅 ABSEN BY MEGUMIN]\n\nJangan lupa absen yaw guys!\nLink absen : https://linktr.ee/aryahmph\n\n^Megumin~"
		participants := handler.getGroupParticipantsJIDs(message.Info.RemoteJid)
		webMessageInfo := handler.generateWebMessageInfo(message, text, participants)

		_, err = handler.conn.Send(webMessageInfo)
		exception.LogIfError(err)
	}
}

func (handler *WhatsappHandlerImpl) getGroupParticipantsJIDs(remoteJid string) []string {
	metaData, err := handler.conn.GetGroupMetaData(remoteJid)
	exception.LogIfError(err)

	var groupMetaData entity.Group
	err = json.Unmarshal([]byte(<-metaData), &groupMetaData)
	exception.LogIfError(err)

	var participants []string
	for _, participant := range groupMetaData.Participants {
		participants = append(participants,
			strings.Replace(participant.Id, "c.us", "s.whatsapp.net", 1))
	}

	return participants
}

func (handler *WhatsappHandlerImpl) generateWebMessageInfo(message whatsapp.TextMessage, text string, participantsJIDs []string) *proto.WebMessageInfo {
	// Generate ID
	b := make([]byte, 10)
	rand.Read(b)

	isTrue := true
	now := uint64(time.Now().Unix())
	status := proto.WebMessageInfo_PENDING
	id := strings.ToUpper(hex.EncodeToString(b))

	return &proto.WebMessageInfo{
		Key: &proto.MessageKey{
			RemoteJid: &message.Info.RemoteJid,
			Id:        &id,
			FromMe:    &isTrue,
		},
		Status:           &status,
		MessageTimestamp: &now,
		Message: &proto.Message{
			ExtendedTextMessage: &proto.ExtendedTextMessage{
				Text: &text,
				ContextInfo: &proto.ContextInfo{
					MentionedJid: participantsJIDs,
				},
			},
		},
	}
}
