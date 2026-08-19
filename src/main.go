package main

import (
	"aspire/are/main/lib"
	"aspire/are/main/runtime"
)

func main() {
	defer func() {
		if !lib.DEBUG_MODE {
			err := recover()
			if err != nil {
				lib.Println(err)
			}
		}
	}()
	runtime.SetupARE()
	var ARE = runtime.NewRuntime()
	w := ARE.Worker(false)
	w.ExecModule(ARE.ParseProgram("main.as", true, false, false).OwnModule)
	// runtime.REPL()
}
