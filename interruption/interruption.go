package interruption

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

var sigChan = make(chan os.Signal, 1)

func init() {
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
}

func Wait() {
	sig := <-sigChan
	log.Println("[interruption]", sig.String())
}
