# ExecPool

[![License](https://img.shields.io/badge/license-mit-green.svg)](https://github.com/hexdigest/execpool/blob/master/LICENSE)
[![Build](https://github.com/hexdigest/execpool/actions/workflows/go.yml/badge.svg)](https://github.com/hexdigest/execpool/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/hexdigest/execpool/badge.svg?branch=main)](https://coveralls.io/github/hexdigest/execpool?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/hexdigest/execpool)](https://goreportcard.com/report/github.com/hexdigest/execpool)
[![GoDoc](https://godoc.org/github.com/hexdigest/execpool?status.svg)](http://godoc.org/github.com/hexdigest/execpool)
[![Release](https://img.shields.io/github/release/hexdigest/execpool.svg)](https://github.com/hexdigest/execpool/releases/latest)
[![Awesome](https://cdn.rawgit.com/sindresorhus/awesome/d7305f38d29fed78fa85652e3a63e154dd8e8829/media/badge.svg)](https://github.com/avelino/awesome-go#goroutines)

Why? Sometimes when you implement Go services you have to integrate with third party command line tools. 
You can do it by simply using [exec.Command](https://golang.org/pkg/os/exec/#Command) but the problem with this approach is 
that your service has to wait while a new process is being loaded into the memory and started. Sometimes this can drastically 
increase the latency of your service. ExecPool works similarly to [FastCGI](https://en.wikipedia.org/wiki/FastCGI) but it can
wrap any regular process. It spins up a given number of processes in advance and when it's time to handle a request from the
user your service just attaches stdin to an existing process from the pool. Basically the execpool helps you trade memory for latency.

# Usage

```go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"

	"github.com/hexdigest/execpool"
)

func main() {
	cmd := exec.Command("grep", "none")

	//spin up 100 processes of grep
	pool, err := execpool.New(cmd, 100)

	rc := pool.Exec(strings.NewReader("this makes sense\nthis is nonesense"))
	b, err := ioutil.ReadAll(rc)
	if err != nil {
		log.Fatalf("failed to read from stdout: %v", err)
	}

	// this is nonesense
	fmt.Println(string(b))
}
```

# Benchmark

This benchmark compares "standard" approach with exec.Command and execpool.Pool by running
grep 100 times. For heavier processes you can expect a bigger difference.
```
make benchmark
goos: darwin
goarch: amd64
pkg: github.com/hexdigest/execpool
cpu: Intel(R) Core(TM) i7-6820HQ CPU @ 2.70GHz
BenchmarkNew-8               100            941753 ns/op
BenchmarkCmd-8               100           2386990 ns/op
```

