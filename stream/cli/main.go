package main

import (
	"encoding/json"
	"os"

	stream "github.com/dynacrypt/go-bloxroute/stream"
	"github.com/rs/zerolog"
)

var log = zerolog.New(os.Stderr).Level(zerolog.WarnLevel).With().Timestamp().Logger()

func main() {
	accountID := os.Getenv("ACCOUNT_ID")
	if accountID == "" {
		log.Fatal().Msg("ACCOUNT_ID not set in environment!")
	}

	secretHash := os.Getenv("SECRET_HASH")
	if secretHash == "" {
		log.Fatal().Msg("SECRET_HASH not set in environment!")
	}

	url := os.Getenv("WS_URL")

	s, err := stream.NewStream(
		stream.Account(accountID, secretHash),
		stream.URL(url),
		stream.OnConnect(func() { log.Info().Msg("Connected to tx stream") }),
		stream.OnReconnect(func() { log.Info().Msg("Reconnected to tx stream") }),
		stream.OnError(func(err error) { log.Error().Msg(err.Error()) }),
	)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	ch, err := s.Start()

	enc := json.NewEncoder(os.Stdout)
	for tx := range ch {
		if err != enc.Encode(tx) {
			log.Fatal().Msg(err.Error())
		}
	}
}
