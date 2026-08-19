//@ts-nocheck arachnoscript
globalThis.console = {
  assert(value, message, ...optionalParams) {
    if (typeof message != "string" && message != undefined) {
      throw new Error(
        "TypeError",
        "Second argument for 'message' must be a string.",
      );
    }
    if (!value) {
      message ??= "Assertion failed.";
      for (let param of optionalParams) {
        if (typeof param != "string") param = #_to_string(param);
      }
      throw #_sprintf(message, ...optionalParams);
    }
  },
  log(...args) {
    if (#_length(args) == 0) {
      #_fs_write_file(#_stdout, #_new_byte_array("\n"));
      return;
    }
    if (#_length(args) == 1) {
      #_fs_write_file(#_stdout, #_new_byte_array(#_inspect(args[0]) + "\n"));
      return;
    }
    if (typeof args[0] == "string" && #_string_includes(args[0], "%")) {
      for (const key in args) {
        if (typeof args[key] == ("boolean" | "string" | "number")) continue;
        args[key] = #_inspect(args[key]);
      }
      const bytes = #_new_byte_array(#_append(#_sprintf(...args), "\n"));
      #_fs_write_file(#_stdout, bytes);
    } else {
      for (const key in args) {
        if (typeof args[key] != "string") args[key] = #_inspect(args[key]);
        args[key] += " ";
      }
      const bytes = #_new_byte_array(#_append(...args, "\n"));
      #_fs_write_file(#_stdout, bytes);
    }
  },
};
