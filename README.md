# chitian-victor/errgroup
`chitian-victor/errgroup` is a drop-in alternative to Go's wonderful
[`sync/errgroup`](https://pkg.go.dev/golang.org/x/sync/errgroup) but
independent of context.Cancel event. 

There are also some more features as follows:
- catch panic error
- return immediately after error reporting
- but double goroutine

## Usage
The exported API of this package mirrors the `sync/errgroup` package.
The only change needed is the import path of the package, from:
```go
import (
  "golang.org/x/sync/errgroup"
)
```

to

```go
import (
  "github.com/chitian-victor/errgroup"
)
```
Then use in the normal manner. See the [godoc]for more.

```go
g, ctx := errgroup.WithCancel()
// g.SetLimit(10)
g.Go(func() error {
    // do something
    return nil
})

err := g.Wait()
```
