# neilotoole/errgroup
`chitian-victor/errgroup` is a drop-in alternative to Go's wonderful
[`sync/errgroup`](https://pkg.go.dev/golang.org/x/sync/errgroup) but
independent of context.Cancel event. 

There are also some more features as follows:
- catch panic error
- return immediately after error reporting
- but double goroutine
