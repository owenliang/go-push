package go_push

// websocket messgae
type WSMessage struct {
	msgType int
	msgData []byte
}

func BuildWSMessage(msgType int, msgData []byte) (wsMessage *WSMessage) {
	return &WSMessage{
		msgType: msgType,
		msgData: msgData,
	}
}