//@ts-nocheck arachnoscript

/**
 * ArachnoScript Standard Library.
 */

import "numbers.as";
import "math.as";
import "objects.as";
import "errors.as";
import "symbols.as";
import "json.as";
import "console.as";
import "strings.as";

let i = 0

@benchmark for (const _ of #_new_byte_array(1024 ** 2)) {
  if (i % 100 == 0) {
    console.log(i);
  }
  i++
}