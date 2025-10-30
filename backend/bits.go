package main

import "unsafe"

func bool_to_uint32(b bool) uint32 {
	return *(*uint32)(unsafe.Pointer(&b))
}

func clear_bit(n uint32, i int) uint32 {
	mask := ^(uint32(1) << uint32(i))
	n &= mask
	return n
}

func set_bit(n uint32, i int) uint32 {
	n |= 1 << uint32(i)
	return n
}

func get_bit(n uint32, i int) bool {
	mask := (uint32(1) << uint32(i))
	return (n & mask) != 0
}
