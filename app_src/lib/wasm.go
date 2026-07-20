package lib

import (
	"context"
	"errors"
	"fmt"

	"github.com/eliben/watgo"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

type WebAssembly struct {
	ctx     context.Context
	runtime wazero.Runtime
}

func (wasm *WebAssembly) Close() {
	wasm.runtime.Close(wasm.ctx)
}

type Instance struct {
	inst    api.Module
	runtime *WebAssembly
}

// Close closes the resource.
// Note: The context parameter is used for value lookup, such as for logging.
// A canceled or otherwise done context will not prevent Close from succeeding.
func (i Instance) Close() {
	i.inst.Close(i.runtime.ctx)
}

func (i Instance) Name() string {
	return i.inst.Name()
}

func (i Instance) IsClosed() bool {
	return i.inst.IsClosed()
}

type Module struct {
	name    string
	runtime *WebAssembly
	module  wazero.CompiledModule
}

// Close releases all the allocated resources for this CompiledModule.
// Note: It is safe to call Close while having outstanding calls from an api.Module instantiated from this.
func (m Module) Close() {
	m.module.Close(m.runtime.ctx)
}

type Function struct {
	f       api.Function
	runtime *WebAssembly
}

// Call invokes the function with the given parameters and returns any results or an error for any failure looking up or invoking the function.
// Call is not goroutine-safe, therefore it is recommended to create another Function if you want to invoke the same function concurrently.
// On the other hand, sequential invocations of Call is allowed.
// However, this should not be called multiple times until the previous Call returns.
func (m Function) Call(args ...uint64) ([]uint64, error) {
	return m.f.Call(m.runtime.ctx, args...)
}

func WAT2WASM(wat []byte) ([]byte, error) {
	wasm, err := watgo.CompileWATToWASM(wat)
	if err != nil {
		err = errors.New("Wat-to-Wasm Compile Error: " + err.Error())
	}
	return wasm, err
}

func TEST_WASM() {
	wasmString := []byte(`
	  (module
	    (type (func (param i32 i32) (result i32)))
	    (func (type 0)
	      local.get 0
	      local.get 1
	      i32.add)
	    (export "add" (func 0)))
	 `)

	wasmBytes, err := WAT2WASM(wasmString)
	if err != nil {
		fmt.Println(err)
		return
	}

	runtime := NewWASMRuntime()

	compiled, err := WASMCompileModule(runtime, wasmBytes)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer runtime.Close()

	module, err := WASMInstantiateModule(compiled)
	if err != nil {
		fmt.Println(err)
		return
	}

	add, found := GetWASMExport(module, "add")
	if !found {
		fmt.Println("Unexpected behaviour, 'add' function not found.")
		return
	}

	results, err := add.Call(100, 20)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(results[0])
}

func GetWASMExport(instance Instance, name string) (fun Function, found bool) {
	fn := instance.inst.ExportedFunction(name)
	return Function{f: fn, runtime: instance.runtime}, fn != nil
}

// Compiles and instantiates the wasm binary.
// Prefer to use WASMCompileModule then WASMInstantiateModule to Instantiate
func WASMInstantiate(wasm *WebAssembly, src []byte) (Instance, error) {
	inst, err := wasm.runtime.Instantiate(wasm.ctx, src)
	return Instance{
		inst:    inst,
		runtime: wasm,
	}, err
}

// Instantiates a compiled module.
func WASMInstantiateModule(mod Module) (Instance, error) {
	inst, err := mod.runtime.runtime.InstantiateModule(mod.runtime.ctx, mod.module, wazero.NewModuleConfig())
	return Instance{
		inst:    inst,
		runtime: mod.runtime,
	}, err
}

func NewWASMRuntime() *WebAssembly {
	ctx := context.Background()
	runtime := wazero.NewRuntime(ctx)
	return &WebAssembly{
		ctx:     ctx,
		runtime: runtime,
	}
}

// Prefer to use WASMCompileModule then WASMInstantiateModule to Instantiate
func WASMCompileModule(wasm *WebAssembly, wasmBytes []byte) (Module, error) {
	module, err := wasm.runtime.CompileModule(wasm.ctx, wasmBytes)
	return Module{
		name:    module.Name(),
		runtime: wasm,
		module:  module,
	}, err
}
