package bpf

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go  -cflags "-O2 -g -Wall -Werror" -type event -target arm64 BPF c/oom.bpf.c -- -I../headers/aarch64 -I../headers
