//go:build amd64 && !gosec

package vp8

//go:noescape
func computeSAD16x16(a, b []byte) int

//go:noescape
func computeSAD8x8(a, b []byte) int

//go:noescape
func computeSAD4x4(a, b []byte) int
