package main

//go:generate go run github.com/cilium/ebpf/cmd/bpf2gossh  -cflags "-O2 -g -Wall -Werror" -type event -target arm64 bpf c/oom.bpf.c -- -I./headers/aarch64 -I./headers
