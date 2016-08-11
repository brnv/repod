package main

import (
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func getRouterRecovery() gin.HandlerFunc {
	return func(context *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := getStack(3)
				errorf("PANIC: %s\n%s", err, stack)

				return
			}
		}()

		context.Next()
	}
}

func getRouterLogger() gin.HandlerFunc {
	return func(context *gin.Context) {
		start := time.Now()

		// Process request
		context.Next()

		duration := time.Now().Sub(start)

		username, ok := context.Get("username")
		if !ok {
			username = "-"
		}

		infof(
			"%v %v %-4v %v %v %v",
			context.ClientIP(),
			username,
			context.Request.Method,
			context.Request.RequestURI,
			context.Writer.Status(),
			duration,
		)
	}
}

func getStack(skip int) string {
	buffer := make([]byte, 1024)
	for {
		written := runtime.Stack(buffer, true)
		if written < len(buffer) {
			// call stack contains of goroutine number and set of calls
			//   goroutine NN [running]:
			//   github.com/user/project.(*Type).MethodFoo()
			//        path/to/src.go:line
			//   github.com/user/project.MethodBar()
			//        path/to/src.go:line
			// so if we need to skip 2 calls than we must split stack on
			// following parts:
			//   2(call)+2(call path)+1(goroutine header) + 1(callstack)
			// and extract first and last parts of resulting slice
			stack := strings.SplitN(string(buffer[:written]), "\n", skip*2+2)
			return stack[0] + "\n" + stack[skip*2+1]
		}

		buffer = make([]byte, 2*len(buffer))
	}
}
