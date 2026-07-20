package main

import "aspire/are/main/runtime"

func main() {
	var ARE = runtime.NewRuntime(false)
	w := ARE.Worker("main.as", true)
	w.Run()
	// runtime.REPL()
}
