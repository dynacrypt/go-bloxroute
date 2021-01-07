package stream

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jpillora/backoff"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

const (
	CloudWS           = "wss://api.blxrbdn.com/ws"
	EnterpriseCloudWS = "wss://eth.feed.blxrbdn.com:28333"
	DefaultQueueSize  = 100
)

const (
	BackoffMin    = 0 * time.Second
	BackoffMax    = 5 * time.Second
	BackoffFactor = 2
	BackoffJitter = true
)

type StreamOption func(*Stream) error
type ConnectFunc func()
type ErrorFunc func(error)
type ReconnectFunc func()

type Stream struct {
	sync.Mutex
	authHeader  string
	cert        *tls.Certificate
	url         string
	insecure    bool
	backoff     *backoff.Backoff
	txC         chan *Transaction
	onConnect   ConnectFunc
	onError     ErrorFunc
	onReconnect ReconnectFunc
	stopC       chan struct{}
	cache       *cache.Cache
}

func NewStream(opts ...StreamOption) (*Stream, error) {
	var s *Stream
	return s.With(opts...)
}

func (s *Stream) With(opts ...StreamOption) (*Stream, error) {
	if s == nil {
		s = &Stream{}
	}
	for _, f := range opts {
		if err := f(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func Account(accountID, secretHash string) StreamOption {
	return func(s *Stream) error {
		s.authHeader = base64.StdEncoding.EncodeToString([]byte(accountID + ":" + secretHash))
		return nil
	}
}

func CertDir(dir string) StreamOption {
	return func(s *Stream) error {
		dir = filepath.FromSlash(dir)
		cert, err := tls.LoadX509KeyPair(
			filepath.Join(dir, "external_gateway_cert.pem"),
			filepath.Join(dir, "external_gateway_key.pem"),
		)
		if err == nil {
			s.cert = &cert
		}
		return err
	}
}

func Cert(cert tls.Certificate) StreamOption {
	return func(s *Stream) error {
		s.cert = &cert
		return nil
	}
}

func URL(url string) StreamOption {
	return func(s *Stream) error {
		s.url = url
		return nil
	}
}

func Insecure() StreamOption {
	return func(s *Stream) error {
		s.insecure = true
		return nil
	}
}

func Backoff(b *backoff.Backoff) StreamOption {
	return func(s *Stream) error {
		s.backoff = b
		return nil
	}
}

func Chan(ch chan *Transaction) StreamOption {
	return func(s *Stream) error {
		s.txC = ch
		return nil
	}
}

func OnConnect(f ConnectFunc) StreamOption {
	return func(s *Stream) error {
		s.onConnect = f
		return nil
	}
}

func OnError(f ErrorFunc) StreamOption {
	return func(s *Stream) error {
		s.onError = f
		return nil
	}
}
func OnReconnect(f ReconnectFunc) StreamOption {
	return func(s *Stream) error {
		s.onReconnect = f
		return nil
	}
}

func (s *Stream) Start() (chan *Transaction, error) {
	dialer := websocket.DefaultDialer
	url := s.url
	header := http.Header{}
	txC := s.txC
	if txC == nil {
		txC = make(chan *Transaction, DefaultQueueSize)
	}
	bo := s.backoff
	if bo == nil {
		bo = &backoff.Backoff{
			Min:    BackoffMin,
			Max:    BackoffMax,
			Factor: BackoffFactor,
			Jitter: BackoffJitter,
		}
	}
	s.cache = cache.New(1*time.Minute, 2*time.Minute)

	if s.cert != nil {
		dialer.TLSClientConfig = &tls.Config{
			Certificates:       []tls.Certificate{*s.cert},
			InsecureSkipVerify: s.insecure,
		}
		if url == "" {
			url = EnterpriseCloudWS
		}
	} else if s.authHeader != "" {
		header.Set("Authorization", s.authHeader)
		if url == "" {
			url = CloudWS
		}
	} else {
		return nil, errors.New("no authentication set")
	}

	connect := func() (*websocket.Conn, error) {
		conn, _, err := dialer.Dial(url, header)
		if err != nil {
			return nil, err
		}
		req := `{"id": 1, "method": "subscribe", "params": ["newTxs", {"include": ["tx_contents"]}]}`
		if err = conn.WriteMessage(websocket.TextMessage, []byte(req)); err != nil {
			conn.Close()
			return nil, err
		}
		// TODO: Do proper rpc
		_, msg, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			return nil, err
		} else if bytes.Contains(msg, []byte(`"error"`)) {
			conn.Close()
			return nil, errors.New("subscription failed")
		}
		return conn, nil
	}

	go func() {
		first := true
		for {
			conn, err := connect()
			if err != nil {
				if s.onError != nil {
					s.onError(err)
				}
				d := bo.Duration()
				time.Sleep(d)
				continue
			}
			bo.Reset()

			if first {
				if f := s.onConnect; f != nil {
					f()
					first = false
				}
			} else if f := s.onReconnect; f != nil {
				f()
			}

			err = s.fetchTXs(conn, txC)
			conn.Close()
			if f := s.onError; f != nil {
				f(err)
			}
		}
	}()
	return txC, nil
}

func (s *Stream) fetchTXs(conn *websocket.Conn, txC chan *Transaction) error {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		m := Message{}
		err = json.Unmarshal(msg, &m)
		if err != nil {
			return err
		}
		if m.Params == nil || m.Params.Result == nil || m.Params.Result.TX == nil {
			return errors.New("unexpected message format")
		}
		tx := m.Params.Result.TX

		if err = tx.Validate(); err != nil {
			return errors.Wrap(err, "tx validation failed")
		}

		if _, found := s.cache.Get(tx.Hash); !found {
			s.cache.SetDefault(tx.Hash, true)
			txC <- tx
		}
	}
	return nil
}
