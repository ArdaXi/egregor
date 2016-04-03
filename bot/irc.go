package main

import (
	"log"
	"net"
	"time"

	"github.com/ardaxi/egregor"
	"github.com/sorcix/irc"
)

type Bot struct {
	server  string
	name    string
	channel string
	data    chan *irc.Message
	reader  *irc.Decoder
	writer  *irc.Encoder
	conn    net.Conn
	consul  *egregor.ConsulClient
}

func NewBot(server, name, channel string, consul *egregor.ConsulClient) *Bot {
	return &Bot{
		server:  server,
		name:    name,
		channel: channel,
		data:    make(chan *irc.Message, 10),
		consul:  consul,
	}
}

func (b *Bot) Connect() error {
	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		DualStack: true,
	}
	conn, err := dialer.Dial("tcp", b.server)
	if err != nil {
		return err
	}
	b.conn = conn
	b.reader = irc.NewDecoder(conn)
	b.writer = irc.NewEncoder(conn)

	for _, msg := range b.handshake() {
		if err := b.writer.Encode(msg); err != nil {
			return err
		}
	}
	go b.readLoop()
	return nil
}

func (b *Bot) handshake() []*irc.Message {
	return []*irc.Message{
		&irc.Message{
			Command: irc.NICK,
			Params:  []string{b.name},
		},
		&irc.Message{
			Command:  irc.USER,
			Params:   []string{b.name, "0", "*"},
			Trailing: b.name,
		},
	}
}

func (b *Bot) readLoop() error {
	for {
		b.conn.SetDeadline(time.Now().Add(300 * time.Second))
		msg, err := b.reader.Decode()
		if err != nil {
			return err // TODO: Figure out how to handle error
		}
		b.data <- msg
	}
}

func (b *Bot) HandleLoop() {
	for msg := range b.data {
		log.Println(msg)
		switch msg.Command {
		case irc.PING:
			go PingHandler(b, msg)
		case irc.RPL_WELCOME:
			go WelcomeHandler(b, msg)
		case irc.ERR_NICKNAMEINUSE:
			go NickCollisionHandler(b, msg)
		case irc.PRIVMSG:
			MsgHandler(b, msg)
		}
	}
}
