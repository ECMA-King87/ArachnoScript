//@ts-nocheck arachnoscript
// deno-lint-ignore-file

/**
 ***************************************************
 *****************      A R E      *****************
 *****************   M A C R O S   *****************
 *****************      0.2.5      *****************
 ***************************************************
 */

/**
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
type scope = {};

interface Indexable<T> {
  [index: number]: T;
}

namespace os {
  export type File = {};
  export type FileInfo = {};
}

type WebAssemblyRuntime = {};
type WebAssemblyModule = {};
type WebAssemblyInstance = {};

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
 * Converts the given value to a string.
 * Useful for converting byte arrays (slices) to string values.
 *
 * @param v - The value to convert.
 * @returns The converted string.
 */
declare function string_cast(v: any): string;

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
declare function json_stringify(obj: any): string;
declare function json_parse(str: string): any;

declare function at<T>(v: Indexable<T>, idx: number): T;

/*
 * *****************
 *    WebAssembly
 * *****************
 */
declare function new_wasm_runtime(): raw<WebAssemblyRuntime>;

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

declare function wat_to_wasm(bytes: raw<byte[]>): raw<byte[]>;
