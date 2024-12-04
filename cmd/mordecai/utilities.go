package main

import (
	"fmt"
	"time"
)

func showLoadingAnimation(message string, process func() error) error {
	done := make(chan bool)

	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				fmt.Printf("\r%s %s", frames[i], message)
				i = (i + 1) % len(frames)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	err := process()

	done <- true
	fmt.Print("\r\033[K")

	return err
}
