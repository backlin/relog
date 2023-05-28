package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/backlin/relog"
)

func main() {
	fmt.Print("Starting demo\n\n")

	l := relog.NewLogger(os.Stdout)

	for _, u := range updates {
		time.Sleep(u.t * time.Millisecond)
		if err := l.Log(u.id, u.line); err != nil {
			log.Fatal(err)
		}
	}
	if err := l.Close(); err != nil {
		log.Fatal(err)
	}

	fmt.Print("\nDone\n")
}

var updates = []struct {
	t    time.Duration
	id   string
	line string
}{
	{1, "1", "PENDING   job 1"},
	{1, "2", "PENDING   job 2"},
	{1, "3", "PENDING   job 3"},
	{0, "4", "PENDING   job 4"},
	{0, "5", "PENDING   job 5"},
	{900, "1", "RUNNING   job 1"},
	{200, "3", "RUNNING   job 3"},
	{1300, "1", "SUCCESS   job 1: 2.4s runtime, $6.31 billed"},
	{100, "2", "RUNNING   job 2"},
	{1500, "3", "SUCCESS   job 3: 4.9s runtime, $11.74 billed"},
	{200, "4", "RUNNING   job 4"},
	{200, "2", "FAILURE   job 2: Syntax error"},
	{100, "5", "CANCELLED job 5: job 2 failed"},
	{600, "4", "CANCELLED job 4: job 2 failed"},
}
