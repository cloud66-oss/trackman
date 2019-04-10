package main

import (
	"fmt"
	"time"

	"github.com/cloud66/trackman/cmd"
)

func main() {
	defer func() {
		c := make(chan struct{})
		go func() {
			defer close(c)
			cmd.UpdateDone.Wait()
		}()
		select {
		case <-c:
			return
		case <-time.After(30 * time.Second):
			fmt.Println("Update timed out")
			return
		}
	}()
	cmd.Execute()
}
