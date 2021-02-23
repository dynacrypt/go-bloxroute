package sender

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

const (
	CloudWS = "wss://api.blxrbdn.com/ws"
)

type Option func(*Sender)

func Log(log zerolog.Logger) Option {
	return func(s *Sender) {
		s.log = log
	}
}

func AccountID(id string) Option {
	return func(s *Sender) {
		s.accountID = id
	}
}

func SecretHash(hash string) Option {
	return func(s *Sender) {
		s.secretHash = hash
	}
}

func URL(url string) Option {
	return func(s *Sender) {
		s.url = url
	}
}

type Sender struct {
	log        zerolog.Logger
	accountID  string
	secretHash string
	url        string

	conn *websocket.Conn
}

func NewSender(opts ...Option) (*Sender, error) {
	s := &Sender{
		accountID:  "",
		secretHash: "",
		conn:       nil,
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.accountID == "" {
		return nil, errors.New("account id is unset")
	} else if s.secretHash == "" {
		return nil, errors.New("secret hash is unset")
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = tlsConfig

	authHeader := base64.StdEncoding.EncodeToString([]byte(s.accountID + ":" + s.secretHash))
	url := s.url
	if url == "" {
		url = CloudWS
	}

	conn, _, err := dialer.Dial(url, http.Header{"Authorization": []string{authHeader}})
	if err != nil {
		return nil, err
	}

	s.conn = conn
	return s, nil
}

type request struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      int64                  `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

func (s *Sender) Send(rawtx string) (string, error) {
	req := &request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "blxr_tx",
		Params:  map[string]interface{}{"transaction": rawtx},
	}

	msg, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	if err = s.conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		return "", err
	}

	_, resp, err := s.conn.ReadMessage()
	if err != nil {
		return "", err
	}

	return string(resp), nil
}
