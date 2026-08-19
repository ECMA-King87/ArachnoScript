//@ts-nocheck arachnoscript
globalThis.JSON = {};

/**
 * `JSON.parse(text: string, reviver?: (this: any, key: string, value: any) => any): any`
 * Converts a JavaScript Object Notation (JSON) string into an object.
 * @param text — A valid JSON string.
 * @param reviver — A function that transforms the results. This function is called for each member of the object. If a member contains nested objects, the nested objects are transformed before the parent object is.
 * @throws — {SyntaxError} If text is not valid JSON.
 */
#_define_property(globalThis.JSON, "parse", {
  writable: false,
  configurable: false,
  value: function (text, reviver) {
    if (typeof text != "string") {
      throw new Error(
        "TypeError",
        "JSON.parse: First argument must be a valid JSON string.",
      );
    }
    function revive(obj) {
      switch (typeof obj) {
        case "boolean", "string", "number":
          return #_bind_this(
            reviver,
            obj,
            "",
            obj,
          );
        case "object": {
          const entries = #_entries(obj);
          for (const key in entries) {
            entries[key] = [
              entries[key][0],
              #_bind_this(
                reviver,
                obj,
                entries[key][0],
                entries[key][1],
              ),
            ];
          }
          return #_from_entries(entries);
        }
        case "array": {
          for (const key in obj) {
            obj[key] = #_bind_this(reviver, obj, #_to_string(key), obj[key]);
          }
        }
      }
      return obj;
    }
    try {
      const value = #_json_parse(text);
      if (reviver == undefined) return value;
      if (typeof reviver != "function") {
        throw new Error(
          "TypeError",
          "JSON.parse: Second argument must be a function.",
        );
      }
      return revive(value);
    } catch (error) {
      throw new Error("SyntaxError", error);
    }
  },
});
