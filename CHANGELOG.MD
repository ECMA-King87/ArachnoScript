# Changelog

All notable changes to ArachnoScript are documented in this file.

This repository continues the development of ArachnoScript in a new home. Changes made before this repository was created can be found in the changelog of the previous repository:

<https://github.com/ECMA-King87/ArachnoScript-Runtime-Engine-ARE-/blob/main/CHANGELOG.MD>

## Unreleased

## Added

- New repository for the continued development of ArachnoScript.
- Repository-wide documentation.
- Architecture Decision Records (ADRs).
- Ongoing work on the bytecode compiler and virtual machine.

## Changed

- Development is now tracked from this repository.
- Project structure and documentation have been reorganized.

#s Previous History

The following milestones were completed before this repository was created:

## v0.2.5

### Lexer

- Reworked the lexer.
- Optimized tokenization and token definitions.

### Parser

- Reworked the parser.
- Optimized the AST.
- Added support for the "&", "^", "|", "<<", ">>", "~" and "&^" operators.
- Added getter and setter parsing in object literals.
- Replaced the original variable declaration keywords ("spawn", "immortal", "static") with "let" and "const".
- Added support for default values in object destructuring.

### Runtime

- Improved output formatting.

For changes made before v0.2.5, see the previous repository's changelog.