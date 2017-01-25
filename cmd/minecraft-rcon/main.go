package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/SeerUK/minecraft-rcon/rcon"
)

func main() {
	var host string
	var port int
	var pass string

	flag.StringVar(&host, "host", "127.0.0.1", "A minecraft server hostname / ip address")
	flag.IntVar(&port, "port", 25575, "A minecraft server RCON port <rcon.port>")
	flag.StringVar(&pass, "pass", "", "A minecraft RCON password <rcon.password>")
	flag.Parse()

	command := strings.Join(flag.Args(), " ")

	client, err := rcon.NewClient(host, port, pass)
	if err != nil {
		handleError(err)
	}

	response, err := client.SendCommand(command)
	if err != nil {
		handleError(err)
	}

	if response != "" {
		fmt.Println(response)
	}
}

func handleError(err error) {
	fmt.Println(err)
	os.Exit(1)
}
