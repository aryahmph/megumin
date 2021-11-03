package handler

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	generateWebMessageInfoQuoted(message whatsapp.TextMessage, text string, participantsJIDs []string) *proto.WebMessageInfo
	getGroupParticipantsJIDs(remoteJid string) []string
	generateID() string
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

	matchGroupJID := handler.regexGroupId.MatchString(message.Info.RemoteJid)

	if message.Text == "!absen" && matchGroupJID {
		_, err := handler.conn.Read(message.Info.RemoteJid, message.Info.Id)
		exception.LogIfError(err)

		text := "*[📅📌 ABSEN BY MEGUMIN]*\n\nJangan lupa absen yaw guys!\nLink absen : https://linktr.ee/aryahmph\n\n^Megumin~"
		participants := handler.getGroupParticipantsJIDs(message.Info.RemoteJid)
		webMessageInfo := handler.generateWebMessageInfo(message, text, participants)

		_, err = handler.conn.Send(webMessageInfo)
		exception.LogIfError(err)
	} else if (message.Text == "!everyone" || strings.Contains(message.Text, "!everyone")) &&
		matchGroupJID {
		_, err := handler.conn.Read(message.Info.RemoteJid, message.Info.Id)
		exception.LogIfError(err)

		text := "*Tolong dibaca yaa!*"
		participants := handler.getGroupParticipantsJIDs(message.Info.RemoteJid)
		webMessageInfo := handler.generateWebMessageInfoQuoted(message, text, participants)

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

func (handler *WhatsappHandlerImpl) generateID() string {
	b := make([]byte, 10)
	rand.Read(b)
	return strings.ToUpper(hex.EncodeToString(b))
}

func (handler *WhatsappHandlerImpl) generateWebMessageInfo(message whatsapp.TextMessage,
	text string, participantsJIDs []string) *proto.WebMessageInfo {
	isTrue := true
	now := uint64(time.Now().Unix())
	status := proto.WebMessageInfo_PENDING
	id := handler.generateID()

	fmt.Println("Generated")

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

func (handler *WhatsappHandlerImpl) generateWebMessageInfoQuoted(message whatsapp.TextMessage, text string, participantsJIDs []string) *proto.WebMessageInfo {
	isTrue := true
	now := uint64(time.Now().Unix())
	status := proto.WebMessageInfo_PENDING
	id := handler.generateID()

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
					StanzaId: &message.Info.Id,
					QuotedMessage: &proto.Message{
						Conversation: &message.Text,
					},
					Participant:  &message.Info.SenderJid,
					MentionedJid: participantsJIDs,
				},
			},
		},
	}
}
