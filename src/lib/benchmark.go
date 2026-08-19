package lib

import (
	"os"
	"runtime/pprof"
)

func init() {
	if DEBUG_MODE {
		cpuF, err := os.Create("cpu.prof")
		if err == nil {
			pprof.StartCPUProfile(cpuF)
			defer func() {
				pprof.StopCPUProfile()
				cpuF.Close()
			}()
		}
		defer func() {
			memF, err := os.Create("mem.prof")
			if err == nil {
				pprof.WriteHeapProfile(memF)
				memF.Close()
			}
		}()
	}
}

func b(){
}
