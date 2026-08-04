const E = 2.71828182845904523536028747135266249775724709369995957496696763;
const PI = 3.14159265358979323846264338327950288419716939937510582097494459;
const PHI = 1.61803398874989484820458683436563811772030917980576286213544862;

const SQRT2 = 1.41421356237309504880168872420969807856967187537694807317667974;
const SQRT1_2 = 1 / SQRT2;
const SQRTE = 1.64872127070012814684865078781416357165377610071014801157507931;
const SQRTPi = 1.77245385090551602729816748334114518279754945612238712821380779;
const SQRTPhi =
  1.27201964951406896425242246173749149171560804184009624861664038;

const LN2 = 0.693147180559945309417232121458176568075500134360255254120680009;
const LOG2E = 1 / LN2;
const LN10 = 2.30258509299404568401799145468436420760110148862877297603332790;
const LOG10E = 1 / LN10;

globalThis.Math = {
  E,
  PI,
  PHI,

  SQRT2,
  SQRTE,
  SQRTPi,
  SQRTPhi,

  LN2,
  LOG2E,
  LN10,
  LOG10E,
  SQRT1_2,

  max(...values) {
    let maxN = 0;
    let idx = 0;
    const len = #_length(values);
    while (idx < len) {
      const v = values[idx];
      if (typeof v != "number") {
        throw "Math.max: encountered non-numerical value in arguments.";
      }
      if (v > maxN) maxN = v;
      idx++;
    }
    return maxN;
  },

  abs(x) {
    if (typeof x != "number") throw "Math.abs: argument must be a number.";
    return #_abs(x);
  },

  trunc(x) {
    if (typeof x != "number") throw "Math.trunc: argument must be a number.";
    return #_trunc_num(x);
  },
};
