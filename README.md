# videohead

parsing mp4 and extract meta information.

```
package main

import (
	"fmt"
	"os"
	"github.com/suichu/videohead"
)

func main() {
	f, err := os.Open("/path/to/any.mp4")
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
```
