package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ardaxi/egregor"
	"github.com/ardaxi/egregor/pb"
	"golang.org/x/net/context"
)

var appID string

func main() {
	s, err := egregor.NewServer(&egregor.Config{})
	if err != nil {
		log.Fatalf("Egregor error: %v", err)
	}

	val, err := s.GetKey("dollar/appid")
	if err != nil {
		log.Fatalf("Consul error: %v", err)
	}
	appID = string(val)

	usage := []string{
		"dollar returns the current value of the dollar, expressed in Euro",
	}

	s.HandleFunc("dollar", usage, DollarHandler)
	s.Run()
}

type Result struct {
	Timestamp int
	Rates     map[string]float64
}

func DollarHandler(ctx context.Context, req *pb.CommandRequest) (*pb.CommandResponse, error) {
	resp, err := http.Get("https://openexchangerates.org/api/latest.json?app_id=" + appID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	res := Result{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&res); err != nil {
		return nil, err
	}

	return &pb.CommandResponse{
		Reply: []string{fmt.Sprintf("%v EUR", res.Rates["EUR"])},
	}, nil
}
