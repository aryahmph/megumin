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
	"megumin/service"
	"regexp"
	"strconv"
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
	regexNumber          *regexp.Regexp
	GameService          service.GameService
}

func NewWhatsappHandlerImpl(conn *whatsapp.Conn, lastMessageTimestamp uint64, regexGroupId *regexp.Regexp, regexNumber *regexp.Regexp, gameService service.GameService) *WhatsappHandlerImpl {
	return &WhatsappHandlerImpl{conn: conn, lastMessageTimestamp: lastMessageTimestamp, regexGroupId: regexGroupId, regexNumber: regexNumber, GameService: gameService}
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
	if message.Info.Timestamp < handler.lastMessageTimestamp {
		return
	}

	matchGroupJID := handler.regexGroupId.MatchString(message.Info.RemoteJid)

	if message.Text == "!intro" {
		_, err := handler.conn.Read(message.Info.RemoteJid, message.Info.Id)
		exception.LogIfError(err)

		text := "Halo, kenalin nama aku Megumin\n\nSalam kenal yaa 😊"
		_, err = handler.conn.Send(whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{RemoteJid: message.Info.RemoteJid},
			Text: text,
		})
		exception.LogIfError(err)
	} else if message.Text == "!absen" && matchGroupJID {
		_, err := handler.conn.Read(message.Info.RemoteJid, message.Info.Id)
		exception.LogIfError(err)

		text := "*[📅📌 ABSEN BY MEGUMIN]*\n\nJangan lupa absen yaw guys!\nLink absen : https://linktr.ee/aryahmph\n\n^Megumin~"
		participants := handler.getGroupParticipantsJIDs(message.Info.RemoteJid)
		webMessageInfo := handler.generateWebMessageInfo(message, text, participants)

		_, err = handler.conn.Send(webMessageInfo)
		exception.LogIfError(err)
	} else if (message.Text == "!everyone" || strings.Contains(message.Text, "!everyone")) && matchGroupJID {
		_, err := handler.conn.Read(message.Info.RemoteJid, message.Info.Id)
		exception.LogIfError(err)

		text := "*Tolong dibaca yaa!*"
		participants := handler.getGroupParticipantsJIDs(message.Info.RemoteJid)
		webMessageInfo := handler.generateWebMessageInfoQuoted(message, text, participants)

		_, err = handler.conn.Send(webMessageInfo)
		exception.LogIfError(err)
	} else if !handler.GameService.Valid(message.Info.Timestamp) &&
		strings.Contains(message.Text, "!play") && matchGroupJID {

		_, err := handler.conn.Read(message.Info.RemoteJid, message.Info.Id)
		exception.LogIfError(err)

		text := "*[🎮 GAME BY MEGUMIN]*\n\nPermainan tebak-tebakan angka\nTebak angka dari 1 sampai 10\nWaktunya 20 detik mulai dari sekarang!\n\n*Game Dimulai*"
		handler.GameService.RandomGuestNumber()
		participants := handler.getGroupParticipantsJIDs(message.Info.RemoteJid)
		webMessageInfo := handler.generateWebMessageInfo(message, text, participants)

		time.AfterFunc(21*time.Second, func() {
			if !handler.GameService.Valid(message.Info.Timestamp) {
				_, err = handler.conn.Send(whatsapp.TextMessage{
					Info: whatsapp.MessageInfo{
						RemoteJid: message.Info.RemoteJid,
					},
					Text: "Wwkwkkw kalah kok sama bot\n\n*GAME BERAKHIR*",
				})
				exception.LogIfError(err)
			}
			handler.GameService.End(0)
		})

		_, err = handler.conn.Send(webMessageInfo)
		exception.LogIfError(err)

		//go func(remoteJID string, timestamp uint64) {
		//	time.Sleep(20 * time.Second)
		//	if !handler.GameService.Valid(timestamp) {
		//		_, err = handler.conn.Send(whatsapp.TextMessage{
		//			Info: whatsapp.MessageInfo{
		//				RemoteJid: remoteJID,
		//			},
		//			Text: "Wwkwkkw kalah kok sama bot\n\n*GAME BERAKHIR*",
		//		})
		//		exception.LogIfError(err)
		//	}
		//}(message.Info.RemoteJid, message.Info.Timestamp)
	} else if handler.GameService.Valid(message.Info.Timestamp) && handler.regexNumber.MatchString(message.Text) && matchGroupJID {
		_, err := handler.conn.Read(message.Info.RemoteJid, message.Info.Id)
		exception.LogIfError(err)

		atoi, err := strconv.Atoi(message.Text)
		exception.LogIfError(err)

		randomNumber := handler.GameService.GetNumber()
		if atoi == randomNumber {
			handler.GameService.End(0)
			var participant []string
			participant = append(participant, strings.Replace(message.Info.SenderJid, "c.us", "s.whatsapp.net", 1))
			webMessageInfo := handler.generateWebMessageInfoQuoted(message,
				fmt.Sprintf("Selamat @%s tebakan kamu benar! Congratulations desuu~\n\n*Game Berakhir*",
					strings.Split(participant[0], "@")[0]), participant)

			_, err = handler.conn.Send(webMessageInfo)
			exception.LogIfError(err)
		}
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
