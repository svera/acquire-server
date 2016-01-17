package client

type MessageContent struct {
	Typ string
	Det map[string]string
}

type Message struct {
	Author  *Client
	Content MessageContent
}
