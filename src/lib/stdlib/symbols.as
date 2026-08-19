//@ts-nocheck arachnoscript
function Symbol(description) {
  if (typeof description != "string") {
    throw new Error("TypeError", "Argument must be of type string.");
  }
}

#_define_property(Symbol, "for", {
  writable: false,
  configurable: false,
  value: function (key) {
    if (typeof key != "string") {
      throw new Error("TypeError", "Argument must be of type string.");
    }
    return #_symbol_for(key);
  },
});

#_define_property(Symbol, "keyFor", {
  writable: false,
  configurable: false,
  value: function (sym) {
    if (typeof sym != "symbol") {
      throw new Error("TypeError", "Argument must be of type symbol.");
    }
    return #_key_for_symbol(sym);
  },
});

#_define_property(Symbol, "toPrimitive", {
  writable: false,
  configurable: false,
  value: #_symbol_for("toPrimitive"),
});

globalThis.Symbol = Symbol;
