package main

import (
	"./gbroker"
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)


var client *rpc.Client;

var clog *log.Logger;

func connectClientBroker() *rpc.Client {

	cl, err := rpc.Dial("tcp", "localhost:"+gbroker.BROKER_PORT)
	if err != nil {      log.Fatal(err)    }
	return cl
}
func createSubscriber( ) int {
	var reply gbroker.SubscriberReply
	myByte := []byte(" ")
	var err = client.Call("Listener.Subscribe",myByte,  &reply)
	if err != nil {    log.Fatal(err)  }
	return reply.Id

}


func listenHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	switch r.Method {

		case "POST":{
			if err := r.ParseForm(); err != nil {
				fmt.Fprintf(w, "ParseForm() err: %v", err)
				return
			}
			msg := r.FormValue("msg")
			clog.Printf("[CLIENT] Got a message :%s\n",msg)
			fmt.Fprintf(w, "OK")


		}

		default:{
			fmt.Fprintf(w, "Sorry, only POST methods are supported.")
			}
	}

}

func listenToBroker(subscriberId int){
	http.HandleFunc("/", listenHandler)
	http.ListenAndServe(":"+fmt.Sprintf("%d",gbroker.CLIENT_LISTEN_PORT+subscriberId) , nil)

}

func unsubscribe(id int){


	if(id==-1){return}
	clog.Printf("[CLIENT] Unsubscribing")
	var reply gbroker.Reply

	//b  := make([]byte, 4)
	//binary.BigEndian.PutUint32( b, uint32(id))
	bb:=[]byte (strconv.Itoa(id))


	var err = client.Call("Listener.Unsubscribe",bb,  &reply)
	if err != nil {    log.Fatal(err)  }


}
func exitClient(subscriberId int){
	clog.Println("[CLIENT] Exiting")
	unsubscribe(subscriberId)

}

func clientCommandHelp(){

	fmt.Println("start: /s")
	fmt.Println("quit: /q")

}
func main() {

	clog =  log.New(os.Stdout, "", log.Ltime | log.Lmicroseconds)


	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	var subscriberId  =-1;

	in:=make(chan string)
	go gbroker.ReadFromInput(in)
	clientCommandHelp()

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
							clog.Println("[CLIENT] Starting client");
							client=connectClientBroker()
							subscriberId=createSubscriber( )
							//listen to broker messages
							go listenToBroker(subscriberId)

						}
					case "/h":{

						 clientCommandHelp()
					}
					case "/q":
						{
							exitClient(subscriberId)
							break stdinloop

						}

					}
				}

			case  <-sigs:{
				exitClient(subscriberId)
				break stdinloop
			}
			}


		}


}

