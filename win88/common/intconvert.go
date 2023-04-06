package common

import (
	"math"
)

/*
  整型安全转换
*/

const (
	// MaxInt32 int32最大值
	MaxInt32 int32 = math.MaxInt32
	// MinInt32 int32最小值
	MinInt32 = math.MinInt32
)

const (
	// MaxUInt32 uint32最大值
	MaxUInt32 uint32 = math.MaxUint32
)

const (
	// MaxInt64 int64最大值
	MaxInt64 int64 = math.MaxInt64
	// MinInt64 int64最小值
	MinInt64 = math.MinInt64
)

const (
	// MaxUInt64 uint64最大值
	MaxUInt64 uint64 = math.MaxUint64
)

// MakeU64 将两个uint32拼凑为uint64 lo为低字节 hi为高字节
func MakeU64(lo, hi uint32) uint64 {
	return uint64(hi)<<32 | uint64(lo)
}

// LowU32 返回uint64低字节(后uint32)
func LowU32(n uint64) uint32 {
	return uint32(n & 0xffffffff)
}

// HighU32 返回uint64高字节(前uint32)
func HighU32(n uint64) uint32 {
	return uint32(n >> 32)
}

// LowAndHighUI32 分别返回uint64低字节,高字节
func LowAndHighUI32(n uint64) (uint32, uint32) {
	return uint32(n & 0xffffffff), uint32(n >> 32)
}

// MakeI64 将两个int32拼凑为int64 lo为低字节 hi为高字节
func MakeI64(lo, hi int32) int64 {
	return int64(hi)<<32 | int64(lo)
}

// LowI32 返回int64低字节(后int32)
func LowI32(n int64) int32 {
	return int32(n & 0xffffffff)
}

// HighI32 返回int64高字节(前int32)
func HighI32(n int64) int32 {
	return int32(n >> 32)
}

// LowAndHighI32 分别返回int64低字节,高字节
func LowAndHighI32(n int64) (int32, int32) {
	return int32(n & 0xffffffff), int32(n >> 32)
}

// LowAndHighI32 分别返回int64低字节,高字节
func LowAndHighI64(n int64) (int64, int64) {
	return n & 0xffffffff, n >> 32
}
