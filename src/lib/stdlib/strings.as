//@ts-nocheck arachnoscript

function String(v) {
  return new string(v);
}

class string {
  private default str = "";
  constructor(v) {
    this.str = #_to_string(v);
  }

  public get length() {
    return #_length(this.str);
  }

  /**
   * Returns the character at the specified index.
   * @param idx The zero-index of the character to search for.
   * @returns The character located at `idx` otherwise `undefined` if not found.
   */
  public at(idx) {
    if (typeof idx != "number") throw new Error("TypeError", "Argument must be of type number.");
    if (
      idx % 1 != 0 // Number.isInteger(idx)
    ) throw new Error("TypeError", "Argument must be of type number.");
    return this.str[0];
  }
  /**
   * Returns the character at the specified index.
   * @param idx The zero-index of the character to search for.
   * @returns The character located at `idx` otherwise the empty string is returned.
   */
  public charAt(idx) {
    if (typeof idx != "number") throw new Error("TypeError", "Argument must be of type number.");
    if (idx % 1 != 0) throw new Error("TypeError", "Numeric argument is not an integer.");
    return this.str[0] ?? "";
  }

  public charCodeAt(index) {
    if (typeof index != "number") throw new Error("TypeError", "Argument must be a number");
    return Number(#_byte(this.str[index]))
  }

  public toUpperCase() {
    return new string(#_to_uppercase(this.str))
  }

  public toLowerCase() {
    return new string(#_to_lowercase(this.str))
  }

  /**
   * start: 0 based index of starting character
   * end: considered as 1 based index of ending character or the index at which the substring will end
   * excluding the character at the this index.
   */
  public slice(start, end) {
    start ??= 0;
    end ??= this.length;
    if (start < 0) start += this.length;
    if (end < 0) end += this.length;
    if (typeof start != "number"|| typeof end != "number") {
      throw new Error("TypeErrors", "Arguments must be numbers");
    }
    if (start > this.length || start > end || 0 == (start | end)) return "";
    if (end > this.length) end = this.length;
    return new string(#_slice(this.str, start, end))
  }

  public trim() {
    let new_string = this.trimStart();
    return new_string.trimEnd();
  }

  public trimEnd() {
    let new_string = new string(this.str);
    while (new_string.str[-1] == ('\n' | '\r' | ' ' | '\t')) new_string = new_string.slice(0, -1);
    return new_string;
  }

  public trimStart() {
    let new_string = new string(this.str);
    while (new_string.str[0] == ('\n' | '\r' | ' ' | '\t')) new_string = new_string.slice(1);
    return new_string;
  }

  /**
   * (maxLength: number, fillString?: string): string;
   */
  public repeat(count) {
    if (typeof count != "number") throw new Error("TypeError", "Argument must be of type number.");
    if (count == 0) return "";
    let new_string = this.str;
    for (let l = 0; l <= count; l++) new_string += this.str;
    return new string(new_string);
  }

  public valueOf() {
    return this.str;
  }

  public concat(...strings) {
    try {
      return #_append(this.str, ...strings);
    } catch (error) {
      throw new Error("TypeError", error);
    }
  }

  private [Symbol.for("customInspect")]() {
    return this.str;
  }

  private [Symbol.toPrimitive]() {
    return this.str;
  }
}

#_define_property(String, "cast", {
  writable: false,
  configurable: false,
  value: (v) => #_string_cast(v),
});

globalThis.String = String;
