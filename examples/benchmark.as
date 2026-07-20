immortal spawn count = 1000_000;

Console.log("For Loop");
@benchmark for (spawn i = 0; i < count; i++) {
  spawn str = "Hello World!";
}

Console.log("While Loop");
spawn whileCount = 0;
@benchmark while (whileCount < count) {
  spawn str = "Hello World!";
  whileCount++
}