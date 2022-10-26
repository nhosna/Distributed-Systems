
package main

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"./gbroker"
)



var server *rpc.Client;
var slog *log.Logger;


func   connectServerBroker() *rpc.Client {

	serv, err := rpc.Dial("tcp", "localhost:12345")
	if err != nil {      log.Fatal(err)    }

	return serv
}
func printStatus(status  gbroker.PublishStatus,msg string){
	slog.Printf("[SERVER] Received status for message %s\n", msg);

	for i, val := range status.Status {
		slog.Printf("[SERVER] Receive status of client %s is %s\n",i,val );


	}


}

func publish(msg string, ongoingPublishes *sync.WaitGroup ) {

	myByte := []byte(msg)
	var reply * gbroker.PublishStatus
	var err = server.Call("Listener.Publish",myByte,  &reply)
	if err != nil {        log.Fatal(err)      }
	printStatus(*reply,msg)
	ongoingPublishes.Done()


}


func exitServer(ongoingPublishes *sync.WaitGroup){
	slog.Println("[Server] Exiting")
	ongoingPublishes.Wait()
}
func serverCommandHelp(){

	fmt.Println("publish: /p [s|a] <MSG>")
	fmt.Println("start: /s")
	fmt.Println("quit: /q")

}


func main() {
	slog =  log.New(os.Stdout, "", log.Ltime | log.Lmicroseconds)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	in:=make(chan string)
	go gbroker.ReadFromInput(in)

	serverCommandHelp()
	var ongoingPublishes sync.WaitGroup



stdinloop:
	for {
		select {
		case stdin, ok := <-in:
			if !ok {
				break  stdinloop
			} else {

				words := strings.Fields(stdin)

				switchloop:
				switch words[0] {
				case "/s":
					{
						slog.Println("[SERVER] Starting server");
						server=connectServerBroker()

					}
				case "/p"  :
					{
						if(len(words)!=3){
							break switchloop
						}
						slog.Println("[SERVER] Publishing");

						switch words[1]{

						case "s":{
							ongoingPublishes.Add(1)
							publish(words[2],&ongoingPublishes)

						}
						case "a":{
							ongoingPublishes.Add(1)

							go publish(words[2],&ongoingPublishes)
						}
						default:
							{}

						}

					}
				case "/help":{

					serverCommandHelp()
				}
				case "/quit":
					{
						exitServer(&ongoingPublishes)
						break stdinloop

					}

				}
			}

		case  <-sigs:{
			exitServer(&ongoingPublishes)
			break stdinloop
		}
		}


	}



}
