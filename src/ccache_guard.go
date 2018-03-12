package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

func launch(stdout, stderr io.Writer) {
	cmd := exec.Command("bin/ccache")
	cmd.Stdout, cmd.Stderr = stdout, stderr
	if err := cmd.Start(); err != nil {
		time.Sleep(time.Second * 5)
		fmt.Println(time.Now().UTC(), " - start ccache error: ", err)
		return
	}
	if err := cmd.Wait(); err != nil {
		time.Sleep(time.Second)
		fmt.Println(time.Now().UTC(), " - ccache down, error: ", err)
		return
	}
}

func main() {
	ccacheOutput := "/dev/null"
	if len(os.Args) > 1 {
		ccacheOutput = os.Args[1]
	}
	out, err := os.OpenFile(ccacheOutput, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	for {
		if len(os.Args) > 1 {
			bak, err := os.Create(ccacheOutput + ".bak")
			if err != nil {
				out.Seek(0, os.SEEK_SET)
				io.Copy(bak, out)
				bak.Close()
			}
		}
		out.Truncate(0)
		launch(out, out)
	}
}
