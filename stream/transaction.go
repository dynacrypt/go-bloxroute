package stream

import (
	"errors"
)

type Transaction struct {
	Hash     string `json:"hash"`
	From     string `json:"from"`
	To       string `json:"to"`
	Gas      string `json:"gas"`
	GasPrice string `json:"gasPrice"`
	Input    string `json:"input"`
	Value    string `json:"value"`
	Nonce    string `json:"nonce"`
}

func (tx *Transaction) Validate() error {
	switch {
	case tx == nil:
		return errors.New("is nil")
	case tx.Hash == "":
		return errors.New("hash is unset")
	case tx.From == "":
		return errors.New("from is unset")
	case tx.To == "":
		return errors.New("to is unset")
	case tx.Gas == "":
		return errors.New("gas is unset")
	case tx.GasPrice == "":
		return errors.New("gasPrice is unset")
		/*
			case tx.intput == "":
				return errors.New("input is unset")
			case tx.value == "":
				return errors.New("value is unset")
		*/
	case tx.Nonce == "":
		return errors.New("nonce is unset")
	}
	return nil
}
