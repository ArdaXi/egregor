package main

import (
	"sync"

	"github.com/ardaxi/egregor/pb"
)

type receivers struct {
	sync.RWMutex
	m map[int]chan *pb.LogMessage
}

func (r *receivers) Add(c chan *pb.LogMessage) int {
	r.Lock()
	id := len(r.m)
	r.m[id] = c
	r.Unlock()
	return id
}

func (r *receivers) Delete(id int) {
	r.Lock()
	delete(r.m, id)
	r.Unlock()
}

type logServer struct {
	in        chan *pb.LogMessage
	receivers receivers
}

func (s *logServer) Log(stream pb.Log_LogServer) error {
	c := make(chan *pb.LogMessage, 100)
	id := s.receivers.Add(c)
	defer s.receivers.Delete(id)

	for msg := range c {
		err := stream.Send(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *logServer) SendMessage(msg *pb.LogMessage) {
	s.in <- msg
}

func (s *logServer) messageLoop() {
	for msg := range s.in {
		s.receivers.RLock()
		for _, c := range s.receivers.m {
			c <- msg
		}
		s.receivers.RUnlock()
	}
}

func newLogServer() *logServer {
	in := make(chan *pb.LogMessage, 100)
	s := &logServer{in: in}
	go s.messageLoop()
	return s
}
