package config

import "time"

type Config interface {
	Bool(k string, def bool) bool
	Duration(k string, def time.Duration) time.Duration
	String(k string, def string) string
	Int(k string, def int) int
	Int32(k string, def int32) int32
	Int64(k string, def int64) int64
	Uint(k string, def uint) uint
	Uint8(k string, def uint8) uint8
	Uint16(k string, def uint16) uint16
	Uint32(k string, def uint32) uint32
	Uint64(k string, def uint64) uint64
	Float64(k string, def float64) float64
	IntSlice(k string, def []int) []int
	StringSlice(k string, def []string) []string
	Unmarshal(k string, val any) error
	Contains(k string) bool
}
