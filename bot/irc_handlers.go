package main

import (
	"log"
	"strings"
	"time"

	"github.com/ardaxi/egregor/pb"
	"github.com/sorcix/irc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func PingHandler(b *Bot, m *irc.Message) {
	b.writer.Encode(&irc.Message{
		Command:  irc.PONG,
		Params:   m.Params,
		Trailing: m.Trailing,
	})
}

func WelcomeHandler(b *Bot, m *irc.Message) {
	b.writer.Encode(&irc.Message{
		Command: irc.JOIN,
		Params:  []string{b.channel},
	})
}

func NickCollisionHandler(b *Bot, m *irc.Message) {
	nick := m.Params[1]
	b.name = nick + "_"
	b.writer.Encode(&irc.Message{
		Command: irc.NICK,
		Params:  []string{b.name},
	})
}

func MsgHandler(b *Bot, m *irc.Message) {
	nick := m.Prefix.Name
	channel := m.Params[0]
	msg := m.Trailing
	go Log(b, nick, channel, msg, time.Now())
	args := strings.Fields(msg)
	if strings.TrimRight(args[0], ",:") == b.name {
		go CommandHandler(b, nick, channel, args)
	}
}

func Log(b *Bot, nick, channel, body string, stamp time.Time) {
	msg := &pb.Message{
		Nick:    nick,
		Channel: channel,
		Body:    body,
		Time:    stamp.Unix(),
	}
	b.log <- msg
}

func CommandHandler(b *Bot, nick, channel string, args []string) {
	addr, err := b.consul.GetServiceAddr(args[1])
	if err != nil {
		log.Println(err)
		b.writer.Encode(&irc.Message{
			Command:  irc.PRIVMSG,
			Params:   []string{channel},
			Trailing: "No such command.",
		})
		return
	}

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		// TODO: retry mechanics
		return
	}
	defer conn.Close()

	c := pb.NewCommandClient(conn)

	r, err := c.DoCommand(context.Background(), &pb.CommandRequest{
		Nick:    nick,
		Command: args[1],
		Args:    args[2:],
	})
	if err != nil {
		// TODO: retry mechanics
		return
	}

	for _, msg := range r.Reply {
		b.writer.Encode(&irc.Message{
			Command:  irc.PRIVMSG,
			Params:   []string{channel},
			Trailing: msg,
		})
	}
}
