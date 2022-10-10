package main

import "webscrapper/http_server"

func main() {
	server := http_server.InitServer("httpApi")
	server.StartServe()

}
