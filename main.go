package main

import (
	"fmt"
	pb "local/djob/message"
	"local/djob/scheduler"
)

func main() {
	ch := make(chan *pb.Job)
	s := scheduler.New(ch)
	s.Start()
	err := s.AddJob(&pb.Job{Schedule: "@at 2017-07-62T11:15:00+08:00", Name: "once"})
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("add 1")
	s.AddJob(&pb.Job{Schedule: "@every 10s", Name: "every10s"})
	fmt.Println("add 2")
	s.AddJob(&pb.Job{Schedule: "@minutely", Name: "everyminutely"})
	fmt.Println("add 3")
	s.AddJob(&pb.Job{Schedule: "1-10 * * * * *", Name: "cron"})
	fmt.Println("add 4")
	for {
		job := <-ch
		fmt.Printf("Job Name: %s, spec: %s,\n", job.Name, job.Schedule)
	}
}
