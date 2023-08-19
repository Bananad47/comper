package main

import (
	"io"
	"log"
	"time"

	"github.com/emersion/go-message/mail"

	"github.com/emersion/go-imap"

	"github.com/emersion/go-imap/client"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Addr        string `env:"ADDR" env-required`
	Login       string `env:"LOGIN" env-required`
	Password    string `env:"PASSWORD" env-required`
	LastMessage uint32 `env:"LASTMESSAGE"`
}

type MessageInfo struct {
	Body       string
	SenderMail string
	SenderName string
	Date       string
	Subject    string
}

var cfg Config
var lastMessage uint32

// считываем конфиг из .env перед запуском
func init() {
	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		log.Fatal(err)
	}
	lastMessage = cfg.LastMessage
}

func main() {
	log.Println("Config:", cfg)
	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS(cfg.Addr, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login(cfg.Login, cfg.Password); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// Select INBOX

	for {
		mbox, err := c.Select("INBOX", false)
		if err != nil {
			log.Fatal(err)
		}
		if mbox.Messages == lastMessage {
			log.Println("No new messages")
			time.Sleep(10 * time.Second)
			continue
		}
		seqset := new(imap.SeqSet)
		seqset.AddRange(lastMessage+1, mbox.Messages)
		log.Println(lastMessage, mbox.Messages)
		lastMessage = mbox.Messages
		messages := make(chan *imap.Message, 10)
		done := make(chan error, 1)
		go func() {
			done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchBody, imap.FetchEnvelope, imap.FetchItem("BODY.PEEK[]")}, messages)
		}()
		for msg := range messages {
			log.Println("New message", msg.Envelope.Subject)

			for _, addr := range msg.Envelope.From {
				log.Println(*addr)
			}
//			body, err := GetMessageBody(msg)
//			if err != nil {
//				log.Fatal(err)
//			}
//			log.Println(body)
		}
		if err := <-done; err != nil {
			log.Fatal(err)
		}
	}
}

func GetMessageBody(msg *imap.Message) (string, error) {
	var res string
	//считывае тело сообщения
	for _, literal := range msg.Body {
		mailreader, err := mail.CreateReader(literal)
		if err != nil {
			return "", err
		}
		//получаем часть письма, которая содержит тело
		p, err := mailreader.NextPart()
		if err != nil {
			return "", err
		}
		temp, err := io.ReadAll(p.Body)
		if err != nil {
			return "", err
		}
		res = string(temp)
	}
	return res, nil
}
