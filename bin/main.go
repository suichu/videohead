package main

import (
	"fmt"
	"os"

	"github.com/suichu/videohead"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("videohead [filename]")
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer f.Close()

	h, err := videohead.Decode(f)
	if err != nil {
		panic(err)
	}
	fmt.Printf("width: %dpx height: %dpx\n", h.Size.X, h.Size.Y)
}
