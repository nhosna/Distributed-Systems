package gbroker

import (
	"container/list"
	fmt "fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)


func NewBroker(buffsize int ) *Broker {
	m:= &Broker{
		PublishChannel:    make(chan Msg,buffsize),     //channel that publishers publish to
		BufferNotFullCond: sync.NewCond(&sync.Mutex{}), //used to notify publisher waiting for buffer to become empty
		log:               log.New(os.Stdout, "", log.Ltime | log.Lmicroseconds),
		SubscriberIds:     list.New(), //list of subscirber ids

	}

	return m
}


func (b *Broker) Post(msg string, id int) string {
	port:=fmt.Sprintf("%d", CLIENT_LISTEN_PORT+id )


	data := url.Values {
		"msg": {msg},
	}

	resp, err := http.PostForm("http://127.0.0.1:"+port+"/", data)

	if err != nil {
		panic(err)
	}

	return resp.Status



}


func (b *Broker) Queue(quit chan os.Signal) {



	for{
		select {
		case msg,ok  := <-b.PublishChannel: {

			time.Sleep(10 * time.Second)

			if(!ok){
				b.log.Printf("[BROKER] Subscriber channel is closed\n");
				return

			}
			b.BufferNotFullCond.L.Lock()
			b.BufferNotFullCond.Signal()
			b.BufferNotFullCond.L.Unlock()

			if (msg.Data ==""){
				continue
			}

			b.log.Printf( "[BROKER] Sending %s to %d subscribers\n",msg.Data,b.SubscriberIds.Len());



			var status map[string]string
			status=make(map[string]string);
			var e *list.Element
			for e  = b.SubscriberIds.Front(); e != nil; e = e.Next(){
				id:= e.Value.(int)
				status[strconv.Itoa(id)]=b.Post(msg.Data,id)

			}
			msg.ReceiveStatus <-status
		}

		case <-quit:{
			b.log.Printf( "[BROKER] Exiting\n")
			return

		}
		default:{}

		}

	}

}
func (b *Broker) Close()   {
	b.log.Println("[BROKER] Closing broker channel");

	if(b.PublishChannel !=nil){
		close(b.PublishChannel)

	}



}

//------------------------------Subscriber--------------------------------------
func (b *Broker) Subscribe() ( int) {

	var id=b.NextSubscriberId
	b.NextSubscriberId++
	b.SubscriberIds.PushBack(id)

	b.log.Printf("[BROKER] Subscribed client %d\n",id);
	return  id

}



func (b *Broker) Unsubscribe(  id int )     {
	b.log.Printf("[BROKER] Unsubscribing client %d\n",id);
	var e *list.Element
	for e  = b.SubscriberIds.Front(); e != nil; e = e.Next() {
		if(e.Value==id){break}
	}
	b.SubscriberIds.Remove(e)

}


//------------------------------Publisher--------------------------------------



func (p *Broker) WaitForStatus(msg Msg) map[string]string {

	p.log.Printf( "[SERVER] Waiting for receive status of message '%s' \n",msg.Data)
	status:=<-msg.ReceiveStatus
	p.log.Printf( "[SERVER] Receive status of message '%s' received\n",msg.Data)
	return status
}



func (p *Broker) Publish(mm string ) map[string]string {
	p.log.Printf("[BROKER] Publishing message '%s'\n",mm);

	x :=  Msg{Data: mm, ReceiveStatus:make(chan map[string]string)}
	p.BufferNotFullCond.L.Lock()

forloop:
	for{

		select {
		case p.PublishChannel <- x:
			{

				p.log.Printf("[SERVER] Sent message %s from publisher\n", x.Data);
				p.BufferNotFullCond.L.Unlock()
				break forloop

			}

		default:
			{
				p.log.Printf( "[SERVER] Broker buffer is full\n")
				p.BufferNotFullCond.Wait()
				p.log.Printf("[SERVER] Broker buffer not full anymore\n")


			}

		}

	}

	return p.WaitForStatus(x)

}
