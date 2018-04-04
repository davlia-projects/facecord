package main

import "log"

func main() {
	proxy, err := NewProxyBot()
	if err != nil {
		log.Printf("could not start proxy")
	}
	proxy.Run()
}
