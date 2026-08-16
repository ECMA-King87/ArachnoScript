//@ts-nocheck arachnoscript
// deno-lint-ignore-file

/*
 ***************************************************
 *****************      A R E      *****************
 *****************   M A C R O S   *****************
 *****************      0.2.5      *****************
 ***************************************************
 */

/*
 * ARE runtime macro declarations.
 * Version: 0.2.5
 *
 * This file documents the public macro API exposed by the interpreter. The
 * declarations below describe the functions and types the standard library and
 * runtime expect callers to use when working with strings, files, objects,
 * symbols, JSON, and WebAssembly.
 */

/*
 * *****************
 *     Constants
 * *****************
 */

/**
 * globalThis - access to the global scope.
 */
declare const globalThis: scope;
/**
 * #_stdin - access to the standard input.
 */
declare const stdin: raw<os.File>;
/**
 * #_stdout - access to the standard output.
 */
declare const stdout: raw<os.File>;

type byte = number;
type raw<T> = T;
type scope = intrinsic;
type PropertyKey = string | number | symbol;

interface PropertyDescriptor {
  configurable?: boolean;
  enumerable?: boolean;
  value?: any;
  writable?: boolean;
  get?(): any;
  set?(v: any): void;
}
/**
 * Represents a type that can be indexed.
 */
interface Indexable<T> {
  [index: number]: T;
}
/**
 * Represents a type that can be type casted to `T`.
 */
interface Castable<T> {
}

namespace os {
  export type File = {};
  export type FileInfo = {};
}

interface WebAssemblyRuntime {}
interface WebAssemblyModule {}
interface WebAssemblyInstance {}

declare function assert(v: any): void | never;
declare function inspect(v: any): string;

declare function length(v: any): number;
declare function random(n: number): number;

/*
 * *****************
 *   S T R I N G S
 * *****************
 */

/**
 * Produces the string representation of `v`.
 * This behaves like JavaScript's String() conversion and is useful for
 * debugging and human-readable output.
 *
 * @param v - The value to convert.
 * @returns A string version of `v`.
 */
declare function to_string(v: any): string;
/**
 * Coerces `v` to a string in a runtime-friendly way.
 * This is especially useful when converting byte slices or raw values into
 * textual output.
 *
 * @param v - The value to convert.
 * @returns The converted string.
 */
declare function string_cast(v: any): string;
/**
 * Coerces `v` to a numeric value if possible.
 * This is useful for interpreting byte values and other raw numeric payloads.
 *
 * @param v - The value to convert.
 * @returns A numeric value or NaN when conversion is not valid.
 */
declare function number_cast(v: any): number;

/*
 * *****************
 *  The File System
 * *****************
 */

// 'os' prefix denotes that it takes a path as its argument

/**
 * Opens a file at the given path for reading.
 *
 * The returned handle can be used with the file-system macros below for reading,
 * writing, or inspecting the file.
 *
 * @param path - The path to the file to open.
 * @returns A file handle for the opened file.
 */
declare function os_open_file(path: string): raw<os.File>;

// 'fs' prefix denotes functions that operate on an operating-system file handle.

/**
 * Closes a file handle.
 * After this call, the handle is no longer usable for I/O.
 *
 * @param f - The file to close.
 */
declare function fs_close_file(f: raw<os.File>): void;
/**
 * Reads data from `f` into `buffer`.
 *
 * The function fills as much of `buffer` as possible and returns the number of
 * bytes actually read.
 *
 * @param f - The file to read from.
 * @param buffer - The destination buffer.
 * @returns The number of bytes read.
 */
declare function fs_read_file(f: raw<os.File>, buffer: raw<byte[]>): number;
/**
 * Writes the contents of `buffer` to `f`.
 *
 * @param f - The file to write to.
 * @param buffer - The bytes to write.
 * @returns The number of bytes written.
 */
declare function fs_write_file(f: raw<os.File>, buffer: raw<byte[]>): number;
/**
 * Returns file metadata for `f` as a raw OS file-info object.
 *
 * @param f - The file whose metadata should be inspected.
 * @returns A raw file-info object.
 */
declare function fs_stats(f: raw<os.File>): raw<os.FileInfo>;
/**
 * Returns the size of a file in bytes.
 *
 * @param info - The file metadata returned by `fs_stats`.
 * @returns The number of bytes in the file.
 */
declare function fs_stats_size(info: raw<os.FileInfo>): number;
/**
 * Returns the file name recorded in the file metadata.
 *
 * @param info - The file metadata returned by `fs_stats`.
 * @returns The file name.
 */
declare function fs_stats_name(info: raw<os.FileInfo>): string;
/**
 * Returns the last modification time of the file in Unix-seconds form.
 *
 * @param info - The file metadata returned by `fs_stats`.
 * @returns The file's modification time.
 */
declare function fs_stats_mod_time(info: raw<os.FileInfo>): number;
/**
 * Reports whether the file described by `info` is a directory.
 *
 * @param info - The file metadata returned by `fs_stats`.
 * @returns `true` when the path is a directory, otherwise `false`.
 */
declare function fs_stats_is_dir(info: raw<os.FileInfo>): boolean;

/*
 * *****************
 *   Scope Objects
 * *****************
 */

/**
 * Returns the current execution scope.
 *
 * This exposes the active lexical environment so callers can inspect or modify the
 * current scope state at runtime.
 *
 * @returns The active scope object.
 */
declare function get_context(): scope;
/**
 * Replaces the active execution scope with `ctx`.
 *
 * @param ctx - The scope object to activate.
 */
declare function set_context(ctx: scope): void;

/*
 * ******************
 *    Uint8 Arrays
 * ******************
 */

declare function new_byte_array(): raw<byte[]>;
declare function new_byte_array(length: number): raw<byte[]>;
declare function new_byte_array(str: string): raw<byte[]>;
declare function new_byte_array(...bytes: number[]): raw<byte[]>;

/*
 * *****************
 *      J S O N
 * *****************
 */

/**
 * Serializes an ArachnoScript value into a JSON string.
 *
 * @param obj - The value to stringify.
 * @param indent - When `true`, the output is formatted with indentation.
 * @returns A JSON-formatted string.
 */
declare function json_stringify(obj: any, indent: boolean): string;
/**
 * Parses a JSON string and returns the corresponding ArachnoScript value.
 *
 * @param str - The JSON source text.
 * @returns The reconstructed runtime value.
 */
declare function json_parse(str: string): any;

/**
 * Returns the element at `idx` from an indexable value.
 *
 * @param v - The array-like or string-like value to read from.
 * @param idx - The index to access.
 * @returns The value at that position.
 */
declare function at<E>(v: Indexable<E>, idx: number): E;

/*
 * *****************
 *    WebAssembly
 * *****************
 */

/**
 * Creates a fresh WebAssembly runtime instance.
 *
 * @returns A runtime handle for compiling and instantiating WASM modules.
 */
declare function new_wasm_runtime(): raw<WebAssemblyRuntime>;
/**
 * Compiles a raw WASM binary into a module for the provided runtime.
 *
 * @param runtime - The WebAssembly runtime to use.
 * @param bytes - The binary WASM payload.
 * @returns A compiled WebAssembly module.
 */
declare function wasm_compile(
  runtime: raw<WebAssemblyRuntime>,
  bytes: raw<byte[]>,
): raw<WebAssemblyModule>;

/**
 * Instantiates a compiled WebAssembly module.
 *
 * @param module - The module to instantiate.
 * @returns A live WebAssembly instance.
 */
declare function wasm_instantiate(
  module: raw<WebAssemblyModule>,
): raw<WebAssemblyInstance>;

declare function wasm_instantiate(
  buffer: raw<byte[]>,
  runtime: raw<WebAssemblyRuntime>,
): raw<WebAssemblyInstance>;

declare function wasm_get_export(
  bytes: raw<WebAssemblyInstance>,
  name: string,
): macro;
/**
 * Compiles WebAssembly-Text format to a WASM binary.
 * @param bytes Bytes of WAT to compile.
 */
declare function wat_to_wasm(bytes: raw<byte[]>): raw<byte[]>;

/**
 * Starts the interactive REPL session and blocks until it exits.
 *
 * This is primarily useful for debugging or manually evaluating ArachnoScript
 * expressions at runtime.
 */
declare function repl(): void;
/**
 * Parses a signed integer from a string using the provided radix.
 *
 * Radix values from 2 to 36 are supported, and `0` follows the usual C-style
 * prefixes for binary, octal, hexadecimal, and decimal literals. Underscores are
 * allowed when radix is `0` for readability.
 *
 * If parsing fails, the function returns `NaN`.
 *
 * @param str - The source string to parse.
 * @param radix - The base to use for parsing.
 * @returns The parsed integer value as a number.
 */
declare function parse_int(str: string, radix: number): number;
/**
 * Parses a floating-point number from a string.
 *
 * @param str - The numeric text to parse.
 * @returns The parsed value, or `NaN` if parsing fails.
 */
declare function parse_float(str: string): number;

/**
 * Checks whether `n` is a finite number.
 *
 * @param n - The numeric value to inspect.
 * @returns `true` when the value is finite, otherwise `false`.
 */
declare function is_finite(n: number): boolean;
/**
 * Removes the fractional portion of `n` and keeps the integer part.
 *
 * @param n - The numeric value to truncate.
 * @returns The truncated number.
 */
declare function trunc_num(n: number): number;

/**
 * Returns a stack trace containing up to `n` call frames.
 *
 * @param n - The maximum number of frames to capture.
 * @returns The captured stack trace as a string.
 */
declare function stack_trace(n: number): string;

/**
 * Converts the supplied value to a single byte.
 * This is useful when working with raw binary data and byte-oriented values.
 *
 * @param v - The value to convert.
 * @returns A byte-sized raw value.
 */
declare function byte(v: any): raw<byte>;

declare function define_property<T>(
  o: T,
  p: PropertyKey,
  attributes: PropertyDescriptor,
): T;

/**
 * Coerces `v` into an object form when possible.
 *
 * Macros are intentionally excluded from this conversion and resolve to `null`.
 *
 * @param v - The value to coerce.
 * @returns An object representation of the value.
 */
declare function object_cast(v: any): object;

/**
 * Returns the own property descriptor for a property on an object.
 *
 * This only inspects properties defined directly on the object, not inherited
 * ones from the prototype chain.
 *
 * @param o - The object to inspect.
 * @param p - The property name or key to look up.
 * @returns The property descriptor, or `undefined` when no such property exists.
 */
declare function get_own_prop_descriptor(
  o: any,
  p: PropertyKey,
): PropertyDescriptor | undefined;

/*
 ***************
  S Y M B O L S
 ***************
 */

/**
 * Creates a new unique symbol with the given description.
 *
 * @param description - A human-readable description for the symbol.
 * @returns A new unique symbol value.
 */
declare function new_symbol(description: string): symbol;
/**
 * Returns a symbol from the global registry for `key` if one already exists.
 * Otherwise, it creates and registers a new symbol for that key.
 *
 * @param key - The registry key to look up.
 * @returns The symbol associated with the key.
 */
declare function symbol_for(key: string): symbol;
/**
 * Looks up the key associated with a symbol in the global registry.
 *
 * @param sym - The symbol to reverse-lookup.
 * @returns The registry key, or `undefined` if no matching symbol exists.
 */
declare function key_for_symbol(sym: symbol): string | undefined;

/**
 * Calls `fn` with `thisArg` bound as the function's `this` value.
 *
 * @param fn - The function to call with a bound receiver.
 * @param thisArg - The value to use as `this` when invoking `fn`.
 * @returns The result of invoking the bound function.
 */
declare function bind_this(fn: Function, thisArg: any): any;

/**
 * Converts an object into an array of `[key, value]` entries.
 *
 * @param obj - The source object.
 * @returns An array of key/value pairs.
 */
declare function entries<T, P>(obj: Record<T, P>): [T, P][];
/**
 * Creates an object from an array of key/value entries.
 *
 * @param obj - The entries to build from.
 * @returns A newly constructed object.
 */
declare function from_entries<T, P>(obj: [T, P][]): Record<T, P>;

/**
 * Returns a slice of an indexable value from `start` to `end`.
 *
 * Negative indices are not supported.
 *
 * @param v - The indexable value to slice.
 * @param start - The starting index, inclusive.
 * @param end - The ending index, exclusive.
 * @returns A slice of the original value.
 */
declare function slice<E>(v: Indexable<E>, start: number, end: number): E;
/**
 * Appends new elements to an indexable value and returns the same kind of value.
 *
 * @param obj - The array-like or string-like value to append to.
 * @param elements - Elements to append.
 * @returns The updated value.
 */
declare function append<E>(
  obj: Indexable<E>,
  ...elements: Castable<E>[]
): E;

/**
 * Reports whether `str` contains `sub` as a substring.
 *
 * @param str - The source string.
 * @param sub - The substring to search for.
 * @returns `true` when `sub` is found, otherwise `false`.
 */
declare function string_includes(str: string, sub: string): boolean;

/**
 * Converts a value into its raw underlying runtime representation.
 *
 * @param v - The value to box as a raw value.
 * @returns The raw representation of `v`.
 */
declare function raw_cast(v: any): raw<any>;

/**
 * Retrieves the value stored in an object's default slot.
 */
// declare function default(obj: object): any;

/**
 * Converts a string to uppercase.
 *
 * @param str - The input string.
 * @returns The uppercase version of the string.
 */
declare function to_uppercase(str: string): string;
/**
 * Converts a string to lowercase.
 *
 * @param str - The input string.
 * @returns The lowercase version of the string.
 */
declare function to_lowercase(str: string): string;

/**
 * Appends the elements of `t` to `s`.
 *
 * Both values must be arrays or slices with the same element type.
 *
 * @param s - The destination array.
 * @param t - The array of elements to append.
 * @returns The combined array contents.
 */
declare function append_array<E>(s: raw<E[]>, t: raw<E[]>): string;
