package client

import (
	"github.com/svera/tbg-server/interfaces"
)

type NullClient struct{}

func (n *NullClient) ReadPump(channel interface{}, unregister chan interfaces.Client) {}
func (n *NullClient) WritePump()                                                      {}
func (n *NullClient) Incoming() chan []byte                                           { return make(chan []byte) }
func (n *NullClient) Owner() bool                                                     { return false }
func (n *NullClient) SetOwner(v bool) interfaces.Client                               { return n }
func (n *NullClient) Name() string                                                    { return "null" }
func (n *NullClient) SetName(v string) interfaces.Client                              { return n }
