package models

type Quote struct {
	Symbol   string  `json:"symbol"`
	Exchange string  `json:"exchange"`
	LTP      float64 `json:"ltp"`
	Source   string  `json:"source"`
}
