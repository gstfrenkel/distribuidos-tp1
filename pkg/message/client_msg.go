package message

import (
	"encoding/binary"
	"fmt"
	"net"
	"tp1/pkg/ioutils"
)

type ClientMessage struct {
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

const msgIdSize = 1
const payloadSize = 8

func SendMessage(conn net.Conn, msg ClientMessage) error {
	// Create message
	finalMessage := make([]byte, 0, msgIdSize+payloadSize+len(msg.Data))
	// Append ID
	finalMessage = append(finalMessage, msg.ID)
	lenBytes := make([]byte, payloadSize)
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

func DataCSVReviewsFromBytes(b []byte) (DataCSVReviews, error) {
	var m DataCSVReviews
	return m, fromBytes(b, &m)
}

func (m DataCSVReviews) ToBytes() ([]byte, error) {
	return toBytes(m)
}

func DataCSVGamesFromBytes(b []byte) (DataCSVGames, error) {
	var m DataCSVGames
	return m, fromBytes(b, &m)
}

func (m DataCSVGames) ToBytes() ([]byte, error) {
	return toBytes(m)
}
