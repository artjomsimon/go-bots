package main

import "syscall"

func Wtime_sec() float64 {
	var tv syscall.Timeval
	_ = syscall.Gettimeofday(&tv)
	return float64(tv.Sec) + float64(tv.Usec)/1e6
}

func Wtime_msec() float64 {
	var tv syscall.Timeval
	_ = syscall.Gettimeofday(&tv)
	return float64(tv.Sec)*1e3 + float64(tv.Usec)/1e3
}

func Wtime_usec() float64 {
	var tv syscall.Timeval
	_ = syscall.Gettimeofday(&tv)
	return float64(tv.Sec)*1e6 + float64(tv.Usec)
}
