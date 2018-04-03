package main

import "log"

func main() {
	proxy, err := NewFacebookProxy()
	if err != nil {
		log.Printf("could not start proxy")
	}
	proxy.Run()
}
