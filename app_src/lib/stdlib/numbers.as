//@ts-nocheck arachnoscript
/**
 * An object that represents a number of any kind.
 * All ArachnoScript numbers are 64-bit floating-point numbers.
 * @param {any} v
 */
function Number(v) {
  switch (typeof v) {
    case "number":
      return v;
    case "string":
      return +v;
    default:
      return #_number_cast(v);
  }
}

/**
 * @param {string} s
 * @param {number | undefined} radix
 */
function parseInt(s, radix) {
  if (typeof s != "string") throw "parseInt: first argument must be a string.";
  if (typeof radix != "number") {
    throw "parseInt: second argument must be a number.";
  }
  radix ??= 0;
  return #_parse_int(s, radix);
}

/**
 * @param {string} s
 * @param {number | undefined} radix
 */
function parseFloat(s) {
  if (typeof s != "string") {
    throw "parseFloat: first argument must be a string.";
  }
  return #_parse_float(s);
}

/**
 * @param {number} a
 * @param {number} b
 * @param {number | undefined} relTol Relative Tolerance
 * @param {number | undefined} absTol Absolute Tolerance
 * @returns {boolean}
 */
function approxEquals(a, b, relTol, absTol) {
  if (typeof a != "number") {
    throw "approxEquals: first argument must be a number.";
  }
  if (typeof b != "number") {
    throw "approxEquals: second argument must be a number.";
  }
  relTol ??= Number.DEFAULT_REL_TOL;
  absTol ??= Number.DEFAULT_ABS_TOL;
  if (typeof relTol != "number") {
    throw "approxEquals: third argument must be a number.";
  }
  if (typeof absTol != "number") {
    throw "approxEquals: forth argument must be a number.";
  }
  if (Number.isNaN(a) || Number.isNaN(b)) return false;
  if (a == Infinity || b == Infinity) return a == b;

  const diff = #_abs(a - b);

  if (diff <= absTol) {
    return true;
  }

  return diff <= relTol * Math.max(#_abs(a), #_abs(b));
}

Number.parseInt = parseInt;
Number.parseFloat = parseFloat;
Number.approxEquals = approxEquals;

/**
 * Returns true if the value passed is a safe integer.
 */
Number.isSafeInteger = function (value) {
  if (typeof value != "number") {
    return false;
  }
  return value <= Number.MAX_SAFE_INTEGER && value >= Number.MIN_SAFE_INTEGER;
};
/**
 * Returns a Boolean value that indicates whether a value is the reserved value NaN (not a number).
 */
Number.isNaN = function (number) {
  // if (typeof value != "number") throw "isNaN: argument must be a number.";
  return number != number;
};

globalThis.isNaN = Number.isNaN;

/**
 * Returns true if passed value is finite.
 */
Number.isFinite = function (value) {
  if (typeof value != "number") throw "isFinite: argument must be a number.";
  return #_is_finite(value);
};

globalThis.isFinite = Number.isFinite;

/**
 * Returns true if the value passed is an integer, false otherwise.
 */
Number.isInteger = function (value) {
  if (typeof value != "number") return false;
  return value % 1 == 0;
};

/**
 * The largest number that can be represented in ArachnoScript. Equal to approximately 1.79E+308.
 */
// Number.MAX_VALUE = 0x1p1023 * (1 + (1 - 0x1p-52)); // 1.79769313486231570814527423731704356798070e+308
Number.MAX_VALUE = 2 ** 1023 * (1 + (1 - 2 ** -52));
/**
 * The closest number to zero that can be represented in ArachnoScript. Equal to approximately 5.00E-324
 */
// Number.MIN_VALUE = 0x1p-1022 * 0x1p-52; // 4.9406564584124654417656879286822137236505980e-324
Number.MIN_VALUE = 2 ** -1022 * 2 ** -52;
/**
 * The value of the largest integer n such that n and n + 1 are both exactly representable as a number value.
 * The value of Number.MAX_SAFE_INTEGER is 9007199254740991 2^53 − 1.
 */
// Number.MAX_SAFE_INTEGER = 1 << 62;
Number.MAX_SAFE_INTEGER = 9007199254740991;
/**
 * The value of the smallest integer n such that n and n − 1 are both exactly representable as a Number value.
 * The value of Number.MIN_SAFE_INTEGER is −9007199254740991 (−(2^53 − 1)).
 */
// Number.MIN_SAFE_INTEGER = -1 << 63;
Number.MIN_SAFE_INTEGER = -9007199254740991;
/**
 * A value that is not a number. In equality comparisons, NaN does not equal any value, including itself.
 * To test whether a value is equivalent to NaN, use the isNaN function.
 */
Number.NaN = NaN;
/**
 * A value greater than the largest number that can be represented in ArachnoScript.
 * ArachnoScript displays POSITIVE_INFINITY values as Infinity.
 */
Number.POSITIVE_INFINITY = Infinity;
/**
 * A value that is less than the largest negative number that can be represented in ArachnoScript.
 * ArachnoScript displays NEGATIVE_INFINITY values as -Infinity.
 */
Number.NEGATIVE_INFINITY = -Infinity;

/**
 * The value of Number.EPSILON is the difference between
 * 1 and the smallest value greater than 1
 * that is representable as a Number value, which is approximately: 2.2204460492503130808472633361816e−‍16.
 */
Number.EPSILON = 1 / (1 << 52);

// A loose relative tolerance of 1e-6
Number.LOOSE_REL_TOL = 1e-6;
// An absolute tolerance of 1e-9
Number.LOOSE_ABS_TOL = 1e-9;

// A default relative tolerance of 1e-9
Number.DEFAULT_REL_TOL = 1e-9;
// A default absolute tolerance of 1e-12
Number.DEFAULT_ABS_TOL = 1e-12;

// A strict relative tolerance of 1e-12
Number.STRICT_REL_TOL = 1e-12;
// A strict absolute tolerance of 1e-15
Number.STRICT_ABS_TOL = 1e-15;

globalThis.Number = Number;
