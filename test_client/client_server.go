package main

import "time"

func main() {
	ReadTestConfig()

	for i := 1; i <= Config.ServerConfig.ClientNum; i++ {
		go newClientBot(i).run()
	}
	for true {
		time.Sleep(time.Second)
	}
}
