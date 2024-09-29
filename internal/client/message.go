package client

import (
	"encoding/binary"
	"fmt"
	"net"
	"tp1/pkg/ioutils"
)

type Message struct {
	ID      uint8
	DataLen uint64
	Data    []byte
}

type DataCSVGames struct {
	AppID                   int64
	Name                    string
	ReleaseDate             string
	EstimatedOwners         string
	PeakCCU                 int64
	RequiredAge             int64
	Price                   float64
	DiscountDLCCount        int64
	Blank                   int64
	AboutTheGame            string
	SupportedLanguages      string
	FullAudioLanguages      string
	Reviews                 string
	HeaderImage             string
	Website                 string
	SupportURL              string
	SupportEmail            string
	Windows                 bool
	Mac                     bool
	Linux                   bool
	MetacriticScore         int64
	MetacriticURL           string
	UserScore               int64
	Positive                int64
	Negative                int64
	ScoreRank               float64
	Achievements            int64
	Recommendations         int64
	Notes                   string
	AveragePlaytimeForever  int64
	AveragePlaytimeTwoWeeks int64
	MedianPlaytimeForever   int64
	MedianPlaytimeTwoWeeks  int64
	Developers              string
	Publishers              string
	Categories              string
	Genres                  string
	Tags                    string
	Screenshots             string
	Movies                  string
}

type DataCSVReviews struct {
	AppID       int64
	AppName     string
	ReviewText  string
	ReviewScore int64
	ReviewVotes int64
}

func sendMessage(conn net.Conn, msg Message) error {

	// Create message
	finalMessage := make([]byte, 0, 1+8+len(msg.Data))
	// Append ID
	finalMessage = append(finalMessage, msg.ID)
	lenBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(lenBytes, msg.DataLen)
	// Append length of the payload
	finalMessage = append(finalMessage, lenBytes...)
	// Append payload
	finalMessage = append(finalMessage, msg.Data...)
	if err := ioutils.SendAll(conn, finalMessage); err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}
	return nil
}
