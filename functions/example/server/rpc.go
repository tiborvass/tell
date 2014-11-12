package main

import ("fmt")

//go:generate ./rpc-extract $GOFILE rpc Functions ../client/rpc/rpc.go

type PublicFunctions struct{}

func (f PublicFunctions) F1(b []byte, t bool, i int) (s string, n int) {
	s, n = string(b), i +1
	fmt.Printf("server: F1(%v, %t, %d) -> (%q, %d)\n", b, t, i, s, n)
	return s, n
}

func (f PublicFunctions) F2() {
	fmt.Println("server: F2()")
}