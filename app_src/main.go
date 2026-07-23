package main

import "aspire/are/main/runtime"

func main() {
	runtime.SetupARE()
	var ARE = runtime.NewRuntime()
	w := ARE.Worker(false)
	w.ExecModule(ARE.ParseModule("main.as", true, false, false).OwnModule)
	// runtime.REPL()
}
