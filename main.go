package main

import (
	"fmt"
	"time"

	"github.com/cloud66-oss/trackman/cmd"
	"github.com/cloud66-oss/trackman/utils"
)

func main() {
	defer func() {
		utils.CloseAllFiles()

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
