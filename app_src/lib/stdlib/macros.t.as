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
 * ARE Macros Library
 * Version: 0.2.5
 *
 * This file declares the public macro API for the ARE runtime. It provides
 * type declarations and utility functions used throughout the standard
 * library.
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
 * Returns the string representation of the value `v`.
 * Mimics JavaScript's String() constructor.
 *
 * @param v - The value to convert to a string.
 * @returns A string representation of `v`.
 */
declare function to_string(v: any): string;
/**
 * Converts / coerces the given value to a string.
 * Useful for converting byte arrays (slices) to string values.
 *
 * @param v - The value to convert.
 * @returns The converted string.
 */
declare function string_cast(v: any): string;
/**
 * Converts / coerces the given value to a number.
 * Useful for converting bytes to AS numbers.
 */
declare function number_cast(v: any): number;

/*
 * *****************
 *  The File System
 * *****************
 */

// 'os' prefix denotes that it takes a path as its argument

/**
 * Opens the file at the specified path for reading.
 *
 * If successful, the returned file object can be used for read operations. The
 * underlying file descriptor will be opened in read-only mode (O_RDONLY).
 *
 * @param path - Path to the file to open.
 * @returns A file handle for the opened file.
 */
declare function os_open_file(path: string): raw<os.File>;

// 'fs' prefix denotes that it takes a os file as its argument

/**
 * Closes the file, rendering it unusable for I/O.
 * @param f File to close.
 */
declare function fs_close_file(f: raw<os.File>): void;
/**
 * Reads up to `length(buffer)` bytes from the file `f` into `buffer`.
 *
 * @param f - The file to read from.
 * @param buffer - The buffer to receive the data.
 * @returns The number of bytes read, or a negative value on error.
 */
declare function fs_read_file(f: raw<os.File>, buffer: raw<byte[]>): number;
/**
 * Writes the bytes from `buffer` to the file `f`.
 *
 * This function will fail with an "Access Denied" error if the process lacks
 * write permissions for `f`.
 *
 * @param f - The file to write to.
 * @param buffer - The bytes to write.
 * @returns The number of bytes written.
 */
declare function fs_write_file(f: raw<os.File>, buffer: raw<byte[]>): number;
/**
 * Returns a (raw) object with the description of the file f.
 */
declare function fs_stats(f: raw<os.File>): raw<os.FileInfo>;
/**
 * Returns the file size (number of bytes) specified by the
 * file description in `info`.
 */
declare function fs_stats_size(info: raw<os.FileInfo>): number;
/**
 * Returns the name of a file specified by the
 * file description in `info`.
 */
declare function fs_stats_name(info: raw<os.FileInfo>): string;
/**
 * Returns the last modification time of a file specified by the
 * file description in `info`.
 */
declare function fs_stats_mod_time(info: raw<os.FileInfo>): number;
/**
 * Returns the true if `info` describes a directory and false otherwise.
 */
declare function fs_stats_is_dir(info: raw<os.FileInfo>): boolean;

/*
 * *****************
 *   Scope Objects
 * *****************
 */

/**
 * Returns the current execution scope object.
 *
 * This can be used to inspect or manipulate scope-local values.
 *
 * @returns The current scope object.
 */
declare function get_context(): scope;
/**
 * Sets the current execution scope to `ctx`.
 *
 * @param ctx - The scope object to activate as the current scope.
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
 * Stringifies the given ArachnoScript value into its corresponding
 * JSON representation.
 * @param obj Object to stringify.
 * @param indent Flag reporting whether the output JSON should be formatted or indented.
 */
declare function json_stringify(obj: any, indent: boolean): string;
/**
 * Parses a JSON string and returns the corresponding
 * ArachnoScript value.
 * @param str JSON text.
 */
declare function json_parse(str: string): any;

/**
 * Returns the element at the specified index in the given indexable value.
 * @param v The indexable value.
 * @param idx The index of the element to return.
 * @returns The element at the specified index.
 */
declare function at<E>(v: Indexable<E>, idx: number): E;

/*
 * *****************
 *    WebAssembly
 * *****************
 */

/**
 * Creates a new WASM runtime.
 */
declare function new_wasm_runtime(): raw<WebAssemblyRuntime>;
/**
 * Decodes a WASM binary to WASM module.
 * @param runtime WASM runtime.
 * @param bytes WASM binary.
 */
declare function wasm_compile(
  runtime: raw<WebAssemblyRuntime>,
  bytes: raw<byte[]>,
): raw<WebAssemblyModule>;

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
 * Switches to interactive REPL mode and returns after REPL is closed.
 */
declare function repl(): void;
/**
 * Parses an integer from the given string using the provided base / radix (0, 2 to 36).
 * The string may begin with a leading sign: "+" or "-".
 * If the base argument is 0, the true base is implied by the string's prefix following the sign (if present): 2 for "0b", 8 for "0" or "0o", 16 for "0x", and 10 otherwise.
 * Also, for argument base 0 only, an underscore character is permitted to separate successive digits.
 * `10` is a valid decimal number, `0b1010` is a valid binary number, `0o12` is a valid octal number, and `0xA` is a valid hexadecimal number.
 * `10_000`, `0b1010_0000`, `0o12_34`, and `0xA_B_C` are all valid numbers with underscores.
 * The function returns the parsed number as a 64-bit floating-point value.
 * If the string cannot be parsed as a valid number, it returns NaN.
 * @param str
 * @param radix
 */
declare function parse_int(str: string, radix: number): number;
declare function parse_float(str: string): number;

/**
 * Reports whether `n` is a finite number.
 * @param n
 */
declare function is_finite(n: number): boolean;
/**
 * Returns the integral part of the numeric expression x, removing any fractional digits.
 * If x is already an integer, the result is x.
 * @param n
 */
declare function trunc_num(n: number): number;

/**
 * Returns a stack trace with `n` number frames.
 * @param n Number of frames to collect.
 */
declare function stack_trace(n: number): string;

/**
 * Converts the given value to a byte.
 * This is useful for converting numbers to bytes.
 * @param v The value to convert.
 * @returns The converted byte.
 */
declare function byte(v: any): raw<byte>;

declare function define_property<T>(
  o: T,
  p: PropertyKey,
  attributes: PropertyDescriptor,
): T;

/**
 * Coerces the given value to an object.
 * @returns null if the argument is a macro, a valid object otherwise.
 */
declare function object_cast(v: any): object;

/**
 * Gets the own property descriptor of the specified object.
 * An own property descriptor is one that is defined directly on the object and is not inherited from the object's prototype.
 * @param o Object that contains the property.
 * @param p Name of the property.
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
 * Creates a new unique symbol.
 * @param description Description of symbol.
 */
declare function new_symbol(description: string): symbol;
/**
 * Returns a Symbol object from the global symbol registry matching the given key if found.
 * Otherwise, returns a new symbol with this key.
 * @param key — key to search for.
 */
declare function symbol_for(key: string): symbol;
/**
 * Returns a key from the global symbol registry matching the given Symbol if found.
 * Otherwise, returns undefined.
 * @param sym — Symbol to find the key for.
 */
declare function key_for_symbol(sym: symbol): string | undefined;

/**
 * Calls `fn` and returns the result with `this` set to `thisArg`.
 * @param fn
 * @param thisArg
 */
declare function bind_this(fn: Function, thisArg: any): any;

declare function entries<T, P>(obj: Record<T, P>): [T, P][];
declare function from_entries<T, P>(obj: [T, P][]): Record<T, P>;

/**
 * Returns the elements of the indexable value `v` from `start` to `end`.
 * Negative indices are not allowed.
 */
declare function slice<E>(v: Indexable<E>, start: number, end: number): E;
/**
 * Appends the list of elements to `obj`.
 * The result has the same type as `obj`.
 */
declare function append<E>(
  obj: Indexable<E>,
  ...elements: Castable<E>[]
): E;

declare function string_includes(str: string, sub: string): boolean;

declare function raw_cast(v: any): raw<any>;

/**
 * Retrieves the value in an object's default slot.
 */
// declare function default(obj: object): any;

declare function to_uppercase(str: string): string;
declare function to_lowercase(str: string): string;

/**
 * Append the elements of `t` to `s`.
 * Both must have the same element type.
 */
declare function append_array<E>(s: raw<E[]>, t: raw<E[]>): string;
