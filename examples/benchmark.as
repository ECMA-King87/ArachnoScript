const count = 1000_000;

console.log("For Loop");
@benchmark for (let i = 0; i < count; i++) {
  let str = "Hello World!";
}

console.log("While Loop");
let whileCount = 0;
@benchmark while (whileCount < count) {
  let str = "Hello World!";
  whileCount++
}