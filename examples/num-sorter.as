$ Number Sorter

function isNumber(value) {
  if (typeof value == "string") {
    $ `Number(value)` returns NaN if a valid number does not exist in 'value'.
    $ so if `Number(value)` is not NaN, a number can be extracted from 'value'.
    return !Number.isNaN(Number(value))
  } else if (typeof value == "number") return true;
  return false;
}

function sortNumbers(array) {
  if (typeof array != "array") throw "App crashed: internal error";
  const arraySize = #_length(array);
  let sorted = false, arrayIndex = 0, currentElem, nextElem;

  while (!sorted) {
    arrayIndex = 0; sorted = true;

    while (arrayIndex < arraySize - 1) {
      currentElem = array[arrayIndex];
      nextElem = array[arrayIndex+1];

      if (currentElem > nextElem) {
        array[arrayIndex] = nextElem
        array[arrayIndex+1] = currentElem
        sorted = false
      } else {
        arrayIndex++
      }
    }
  }

  return array
}

function main() {
  let input_str = "";
  let numbers = [];

  print("Welecome to the number sorter! Type `sort` to sort numbers.\r\n");

  while (input_str != "sort") {
    print("Type a number to add to the list: ");
    $ Retrieve the default property of the instance `String` which is a string.
    input_str = #_value(
        String(input()).trim()
      );

    if (isNumber(input_str)) {
      $ Append to the list of numbers the new input.
      #_append(numbers, Number(input_str));
    } else {
      if (input_str != "sort") Console.log(input_str, "is not a valid number.");
    }
  }

  Console.log("Original number list: ", numbers)
  Console.log("Sorted number list: ", sortNumbers(numbers))
}

main()
