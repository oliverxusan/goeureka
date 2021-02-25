package goeureka

type Response struct {
	Status   int         `json:"Status"`
	Code     int         `json:"Code"`
	ErrorMsg string      `json:"ErrorMsg"`
	Data     interface{} `json:"Data"`
}
