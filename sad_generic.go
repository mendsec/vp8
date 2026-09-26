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
