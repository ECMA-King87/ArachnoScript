# ArachnoScript Runtime Changelog

## v0.2.5

### Lexer

- Changed lexer logic.
- Optimized tokens and tokenizer.

### Parser

- Changed parser logic.
- Made many optimizations to AST nodes.
- Enabled parsing of `& ^ | << >> ~ &^` tokens.
- Enabled Getter / Setter parsing in object literals.
- Changed variable declaration keywords `spawn immortal static` to
  `let const`.
- Enabled default value for property in object destructuring.

### Runtime

- Improved formatting of output.

Earlier logs of changes are in the old [repository](https://github.com/ECMA-King87/ArachnoScript-Runtime-Engine-ARE-/blob/main/CHANGELOG.MD).
