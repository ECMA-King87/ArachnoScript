const count = 1000_000;

Console.log("For Loop");
@benchmark for (let i = 0; i < count; i++) {
  let str = "Hello World!";
}

Console.log("While Loop");
let whileCount = 0;
@benchmark while (whileCount < count) {
  let str = "Hello World!";
  whileCount++
}