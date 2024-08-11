package main

import (
	"fmt"
	r "xperim/router"
)

func main() {
	router := r.NewRouter()

	router.GET("/test", func(params map[string]string, env map[string]interface{}) {
		fmt.Println("Hello from the test route")
	})

	handler, params := router.FindHandler("GET", "/test")

	if handler != nil {
		handler(params, nil)
	} else {
		fmt.Println("Handler not found!")
	}
}
