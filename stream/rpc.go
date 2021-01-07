package stream

//	{"jsonrpc": "2.0", "id": null, "method": "subscribe", "params": {"subscription": "c294f047-ec21-43f2-b9da-13e867365f42", "result": {"txContents": {"from": "0x50d6bdfc451314fb162d7d3322bfb4a005cf192f", "gas": "0x7a120", "gasPrice": "0x17bfac7c00", "hash": "0xeef70455ea7118ff6d99883b855d85c045b7f471eab454ea45728b8c4d8b227c", "input": "0x202ee0ed000000000000000000000000000000000000000000000000000000000000211d0000000000000000000000000000000000000000000000000000000002662c5d", "nonce": "0xd0e2", "value": "0x0", "v": "0x25", "r": "0x33b2a65126d2731d700472ee9794b363f14e38bfb1e518229fd7182cbb46371d", "s": "0x262851284702d8756e5036879a1a4808e1bf3954929d1bca1b93c65254f734d6", "to": "0xd286af227b7b0695387e279b9956540818b1dc2a"}}}}
type Message struct {
	Version string  `json:"jsonrpc"`
	Method  string  `json:"method"`
	Params  *Params `json:"params"`
}

type Params struct {
	Subscription string  `json:"subscription"`
	Result       *Result `json:"result"`
}

type Result struct {
	TX *Transaction `json:"txContents"`
}
