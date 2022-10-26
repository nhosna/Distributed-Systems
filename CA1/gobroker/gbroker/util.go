package gbroker


import (
	"bufio"
	"os"
)


func ReadFromInput(in chan string) {

	reader := bufio.NewReader(os.Stdin)
	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			close(in)
			return
		}
		in <- s
	}
}
