//@ts-nocheck arachnoscript
class Object {}

#_define_property(Object, "defineProperty", {
  writable: false,
  configurable: false,
  // If desc == null
  //   o.p == null
  // Else o.p == desc.value
  value: function (o, p, desc) {
    if (typeof o != "object" || o == null) {
      throw "defineProperty: Called on non-object: " + o;
    }
    if (typeof p != "number" && typeof p != "string" && typeof p != "symbol") {
      throw "defineProperty: Invalid property key: " + p;
    }
    if (p in o && !#_get_own_prop_descriptor(o, p).configurable) {
      throw "defineProperty: Cannot redefine property: " + typeof o.p;
    }
    if (typeof desc != "object") {
      throw "defineProperty: Property description must be an object: " + desc;
    }

    if (desc == null) return (0, o.p = null, o);

    if (typeof desc.enumerable != "boolean" && desc.enumerable != undefined) {
      throw "defineProperty: Enumerable descriptor must be of type boolean.";
    }
    if (typeof desc.writable != "boolean" && desc.writable != undefined) {
      throw "defineProperty: Writable descriptor must be of type boolean.";
    }
    if (
      typeof desc.configurable != "boolean" && desc.configurable != undefined
    ) {
      throw "defineProperty: Configurable descriptor must be of type boolean.";
    }

    return #_define_property(o, p, desc);
  },
});
