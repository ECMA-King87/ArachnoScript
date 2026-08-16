//@ts-nocheck arachnoscript
class Error {
  name = "";
  message = "";
  stack = "";
  constructor(name, message) {
    if (typeof name != "string" || typeof message != "string") {
      throw new Error("TypeError", "Both arguments must be of type string.");
    }
    this.name = name;
    this.message = message;
    this.stack = #_stack_trace(10);
  }

  private [Symbol.for("customInspect")]() {
    return this.name + ": " + this.message + "\n" + this.stack;
  }
}

globalThis.Error = Error;
