#!/bin/bash

# Create sad.go
cat << 'GOEOF' > sad.go
package vp8

// Declarations of SAD functions. Implementations are in sad_generic.go or sad_amd64.s.
GOEOF

# Create sad_generic.go
cat << 'GOEOF' > sad_generic.go
//go:build !amd64

package vp8

func computeSAD16x16(a, b []byte) int {
	sad := 0
	for i := 0; i < 256; i++ {
		diff := int(a[i]) - int(b[i])
		mask := diff >> 31
		sad += (diff ^ mask) - mask
	}
	return sad
}

func computeSAD8x8(a, b []byte) int {
	sad := 0
	for i := 0; i < 64; i++ {
		diff := int(a[i]) - int(b[i])
		mask := diff >> 31
		sad += (diff ^ mask) - mask
	}
	return sad
}

func computeSAD4x4(a, b []byte) int {
	sad := 0
	for i := 0; i < 16; i++ {
		diff := int(a[i]) - int(b[i])
		mask := diff >> 31
		sad += (diff ^ mask) - mask
	}
	return sad
}
GOEOF

# Create sad_amd64.go
cat << 'GOEOF' > sad_amd64.go
//go:build amd64

package vp8

//go:noescape
func computeSAD16x16(a, b []byte) int

//go:noescape
func computeSAD8x8(a, b []byte) int

//go:noescape
func computeSAD4x4(a, b []byte) int
GOEOF

# Create sad_amd64.s
cat << 'GOEOF' > sad_amd64.s
#include "textflag.h"

// func computeSAD16x16(a, b []byte) int
TEXT ·computeSAD16x16(SB), NOSPLIT, $0-56
	MOVQ a_base+0(FP), SI
	MOVQ b_base+24(FP), DI
	
	PXOR X2, X2 // accumulator

	// 16 iterations of 16 bytes = 256 bytes
	MOVOU 0(SI), X0
	MOVOU 0(DI), X1
	PSADBW X1, X0
	PADDQ X0, X2

	MOVOU 16(SI), X0
	MOVOU 16(DI), X1
	PSADBW X1, X0
	PADDQ X0, X2

	MOVOU 32(SI), X0
	MOVOU 32(DI), X1
	PSADBW X1, X0
	PADDQ X0, X2

	MOVOU 48(SI), X0
	MOVOU 48(DI), X1
	PSADBW X1, X0
	PADDQ X0, X2

	MOVOU 64(SI), X0
	MOVOU 64(DI), X1
	PSADBW X1, X0
	PADDQ X0, X2

	MOVOU 80(SI), X0
	MOVOU 80(DI), X1
	PSADBW X1, X0
	PADDQ X0, X2

	MOVOU 96(SI), X0
	MOVOU 96(DI), X1
	PSADBW X1, X0
	PADDQ X0, X2

	MOVOU 112(SI), X0
	MOVOU 112(DI), X1
	PSADBW X1, X0
	PADDQ X0, X2

	MOVOU 128(SI), X0
	MOVOU 128(DI), X1
	PSADBW X1, X0
	PADDQ X0, X2

	MOVOU 144(SI), X0
	MOVOU 144(DI), X1
	PSADBW X1, X0
	PADDQ X0, X2

	MOVOU 160(SI), X0
	MOVOU 160(DI), X1
	PSADBW X1, X0
	PADDQ X0, X2

	MOVOU 176(SI), X0
	MOVOU 176(DI), X1
	PSADBW X1, X0
	PADDQ X0, X2

	MOVOU 192(SI), X0
	MOVOU 192(DI), X1
	PSADBW X1, X0
	PADDQ X0, X2

	MOVOU 208(SI), X0
	MOVOU 208(DI), X1
	PSADBW X1, X0
	PADDQ X0, X2

	MOVOU 224(SI), X0
	MOVOU 224(DI), X1
	PSADBW X1, X0
	PADDQ X0, X2

	MOVOU 240(SI), X0
	MOVOU 240(DI), X1
	PSADBW X1, X0
	PADDQ X0, X2

	MOVQ X2, AX
	PSRLDQ $8, X2
	MOVQ X2, CX
	ADDQ CX, AX
	
	MOVQ AX, ret+48(FP)
	RET

// func computeSAD8x8(a, b []byte) int
TEXT ·computeSAD8x8(SB), NOSPLIT, $0-56
	MOVQ a_base+0(FP), SI
	MOVQ b_base+24(FP), DI
	
	PXOR X2, X2 // accumulator

	// 4 iterations of 16 bytes = 64 bytes
	MOVOU 0(SI), X0
	MOVOU 0(DI), X1
	PSADBW X1, X0
	PADDQ X0, X2

	MOVOU 16(SI), X0
	MOVOU 16(DI), X1
	PSADBW X1, X0
	PADDQ X0, X2

	MOVOU 32(SI), X0
	MOVOU 32(DI), X1
	PSADBW X1, X0
	PADDQ X0, X2

	MOVOU 48(SI), X0
	MOVOU 48(DI), X1
	PSADBW X1, X0
	PADDQ X0, X2

	MOVQ X2, AX
	PSRLDQ $8, X2
	MOVQ X2, CX
	ADDQ CX, AX
	
	MOVQ AX, ret+48(FP)
	RET

// func computeSAD4x4(a, b []byte) int
TEXT ·computeSAD4x4(SB), NOSPLIT, $0-56
	MOVQ a_base+0(FP), SI
	MOVQ b_base+24(FP), DI
	
	// 1 iteration of 16 bytes = 16 bytes
	MOVOU 0(SI), X0
	MOVOU 0(DI), X1
	PSADBW X1, X0

	MOVQ X0, AX
	PSRLDQ $8, X0
	MOVQ X0, CX
	ADDQ CX, AX
	
	MOVQ AX, ret+48(FP)
	RET
GOEOF

