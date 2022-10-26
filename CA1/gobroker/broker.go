
package main

import (
	"./gbroker"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)
type Listener int

var broker  *  gbroker.Broker
var blog *log.Logger;



func (l *Listener) Subscribe(line []byte, reply *gbroker.SubscriberReply)  error {
	var  id=  broker.Subscribe()
	*reply = gbroker.SubscriberReply{id}
	return nil
}


func (l *Listener) Unsubscribe(line []byte, reply *gbroker.Reply)  error {


	i,_:=strconv.Atoi(string(line))
	broker.Unsubscribe(i)
	*reply = gbroker.Reply{ }
	return nil
}

func (l *Listener) Publish(line []byte, reply *gbroker.PublishStatus)  error {

	msg:=string(line)

	status:=broker.Publish(msg   ) //TODO is this index correct?
	*reply=gbroker.PublishStatus{status }
	return nil
}


func exitBroker(){

	blog.Println("[BROKER] Exiting")

	if(broker!=nil){
		broker.Close()

	}


}

func brokerCommandHelp(){

	fmt.Println("start:/s [buffsize]")
	fmt.Println("quit:/q")

}
func main() {

	blog =  log.New(os.Stdout, "", log.Ltime | log.Lmicroseconds)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)



	in:=make(chan string)
	go gbroker.ReadFromInput(in)
	brokerCommandHelp()



stdinloop:
	for {
		select {
		case stdin, ok := <-in:
			if !ok {
				break  stdinloop
			} else {

				words := strings.Fields(stdin)

				switch words[0] {
				case "/s":
					{
						if(len(words)!=2){
							continue stdinloop
						}
						buffsize,_:=strconv.Atoi(words[1])

						broker =  gbroker.NewBroker(buffsize  )
						go broker.Queue( sigs)

						blog.Println("[BROKER] Starting broker");

						go func() {
							addy, err := net.ResolveTCPAddr("tcp", "0.0.0.0:"+gbroker.BROKER_PORT)
							if err != nil {     log.Fatal(err)   }
							inbound, err := net.ListenTCP("tcp", addy)
							if err != nil {     log.Fatal(err)   }
							listener := new(Listener)
							rpc.Register(listener)
							rpc.Accept(inbound)
						}()


					}
				case "/h":{

					brokerCommandHelp()
				}

				case "/q":
					{
						exitBroker()
						break stdinloop

					}

				}
			}

		case  <-sigs:{
			exitBroker()
			break stdinloop
		}
		}


	}



}


