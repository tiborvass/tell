package rpc

type Functions struct {
	F1 func(b []byte, t bool, i int) (s string, n int) 
	F2 func() 
}
