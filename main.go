package main

import (
	"fmt"
	"ngrok-client/ngrokc/connection"
)

func main() {


	var a = connection.ControlConnection{}

	a.Init("127.0.0.1", 14443, "asd", "asd")
	a.SetHTTPConfig("", "test", "", 80)
	a.SetHTTPSConfig("", "", "", 443)
	fmt.Println(a.Service())

}