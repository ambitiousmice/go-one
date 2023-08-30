package main

import (
	"github.com/robfig/cron/v3"
	"time"
)

func main() {
	for i := 0; i < 10000; i++ {
		corn := cron.New(cron.WithSeconds())
		corn.Start()
	}
	for true {
		time.Sleep(1 * time.Second)
	}
}
