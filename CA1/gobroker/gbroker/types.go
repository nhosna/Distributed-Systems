package gbroker

import (
	"container/list"
	"log"
	"sync"
)
type Msg struct {
	Data          string
	ReceiveStatus chan map[string]string
}

type Subscriber struct {
	Id     int
}

type Broker struct {
	PublishChannel    chan Msg //the channel for listening to messages from publishers
	SubscriberIds     *list.List
	NextSubscriberId  int
	BufferNotFullCond *sync.Cond
	log               *log.Logger

}

type Reply struct {
}

type PublishStatus struct {
	Status map[string]string
}

type SubscriberReply struct {
	Id int
}



