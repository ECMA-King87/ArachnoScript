const ptr = #_load_shared_lib("msvcrt.dll");

const strlen_symbol = #_ffi_get_symbol(ptr, "strlen", {
  result:     "u64",
  parameters: ["string"],
});

const string = "Super!\x00";
const strPtr = #_unsafe_pointer(string);

#_file_write(
    #_os_stdout(),
    #_to_string(strlen_symbol(strPtr))
  )

#_free_shared_library(ptr);