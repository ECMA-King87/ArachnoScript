/**
 * ArachnoScript Runtime Environment (ARE v0.2.5)
 * - ECMA King
 */

// WebAssembly just dropped in v0.2.5!

// Read from a .wat file, convert to wasm, instantiate, and call an exported function.
const file = #_os_open_file("mod.wat");

// The buffer to hold the contents of the .wat file.
// We use the file size to allocate the buffer.
const watBuffer = #_new_byte_array(
  #_fs_stats_size(
    #_fs_stats(file),
  ),
);

// Read the contents of the .wat file into the buffer.
#_fs_read_file(file, watBuffer);
// Create a new WebAssembly runtime.
const runtime = #_new_wasm_runtime();

// Compile to wasm and instantiate the module.
const instance = #_wasm_instantiate(
  #_wat_to_wasm(watBuffer),
  runtime,
);

// Get the exported function "add" from the wasm instance.
// #_wasm_get_export returns a macro that can be called like a normal function.
// It returns a slice of uint64s.
const add = #_wasm_get_export(instance, "add");

// Call the "add" function with two arguments: 120 and 10.
// Then print the result to the console.
@debug add(1, 10.5);
// Get the first element of the result slice and print it to the console.
@debug #_at(add(1, 10), 0);

// Close the file after we're done with it.
#_fs_close_file(file);
