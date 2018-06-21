package go_push

import "errors"

var (
	ERR_CONNECTION_LOSS = errors.New("CONNECTION IS LOST")

	ERR_SEND_MESSAGE_FULL = errors.New("SEND MESSAGE IS FULL")
)
