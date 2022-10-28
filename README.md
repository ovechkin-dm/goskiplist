Sorted map implementation for go with generics support

```go
require github.com/jet-black/goskiplist v0.0.1
```

```go
package main

import (
	"fmt"
	"github.com/jet-black/goskiplist"
)

func main() {
	m := skiplist.NewMap[int, int](func(a int, b int) int {
		if a < b {
			return 1
		}
		if a > b {
			return -1
		}
		return 0
	})
	m.Add(0, 10)
	x, _ := m.Remove(0)
	fmt.Println(x)
}

```