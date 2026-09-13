‎# ArachnoScript
‎
‎This repository contains the source and the releases of the ArachnoScript Runtime Environment from v0.2.5 and until v0.3.0.
‎This repository also contains documentations of the runtime, language macros, and code examples.
‎
‎## The Language
‎
‎ArachnoScript is a new high level programming language with modern syntax and semantics.
‎It is a backend focused language but also intends to provide built-in frontend capabilities.
‎ArachnoScript uses a syntax that inherits from the JavaScript programming language with decisive differences. The language's semantics are also similar to those of JavaScript but are sometimes more strict to help catch bugs at runtime. The semantics of ArachnoScript stay different from JavaScript's semantics to remove some of the awkwardness that exist in JavaScript.
‎
‎Both the language and its runtime continue to grow and change, moving toward an ambitious version 1 release.
‎
‎## The Runtime
‎
‎The ArachnoScript Runtime Environment (ARE) is the core utility for executing ArachnoScript code.
‎It currently executes ArachnoScript code using an interpreter.
‎The ArachnoScript runtime is written in the Go programming language, speeding up the development of the runtime, also making cross compilation easy.
‎
‎## v0.2.5
‎
‎This version of the runtime marks a significant milestone in its development.
‎Several improvements and performance boosts were made, full rewrites, and language design changes.
‎
‎- Unlike previous versions which required awkward installations of the runtime and the standard library on other platforms besides windows, v0.2.5 has the standard library embedded into it. This means faster startups, a tamper proof standard library, and cleaner installations.
‎- A lot of optimizations were made to the interpreter and the parser, significantly improving execution speed.
‎- The v0.2.5 runtime rivals Deno JavaScript when reading from a single file thousands to millions of times even with V8. The runtime was actually faster, the new string storage in the interpreter and file caching really paid off.
‎- Many language features are now more stable and less error prone.
‎
‎## v0.3.0
‎
‎This planned version of the runtime is already in development and will come with the first stable bytecode interpreter or VM for the language.
‎
‎# Future
‎
‎There are plans for a package manager and many other useful features, but things like language design and the runtime itself have much higher priority.
‎Future releases, v0.3.0 and above, will be partially written in C and Zig, and the rest of the runtime in Go.
‎
‎## Contributing
‎
‎If you would contribute to this project, it would be very helpful and speed up the development process by a lot.
‎This project has got just a single developer, founder, and maintainer.
‎You can contribute through the following ways:
‎
‎- Studying the runtime and testing it for yourself or even using the language and issuing bug reports will help to catch bugs and flaws in both the language, the semantics, and the runtime. You may even help add and remove language features that may or may not suit developers needs. You'll help shape the language's future!
‎- Join one of these social platforms and follow the journey.

‎1. [Discord](https://discord.gg/nQm7v2eXWt)
‎2. [Buzzly](https://buzzly.harislab.tech)
‎3. [Facebook: ECMA King](https://m.facebook.com/ecmaking57)
‎4. [GitHub](https://github.com/ECMA-King87/ArachnoScript)
‎5. [Medium](https://medium.com/@ecmaking87)