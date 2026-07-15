# cli — interactive shell skeleton

An interactive command-line shell scaffold for TinyDB. It reads input in a REPL loop, recognizes a small set of commands, and prints placeholder output. It is a stub and is not connected to the database engine.

## Responsibilities

- Present a prompt and read lines from standard input in a loop.
- Distinguish meta-commands (lines starting with `.`) from SQL statements.
- Perform minimal recognition of `insert` and `select` statements.
- Print placeholder responses (no real execution).

## Key types

- `CLI` — holds the input reader and buffer bookkeeping fields.
- `Statement` — holds the recognized `StatementType`.
- `MetaCommandResult`, `StatementType`, `PrepareResult` — integer enums for command dispatch (e.g. `META_COMMAND_SUCCESS`, `STATEMENT_INSERT`/`STATEMENT_SELECT`, `PREPARE_SUCCESS`).

## Key API

- `func NewCLI() *CLI` — constructs a CLI reading from `os.Stdin`.
- `func (c *CLI) Start()` — runs the REPL: prints help, then loops reading input, dispatching meta-commands, preparing statements, and "executing" them.

## How it fits

Intended as a standalone entry point for interacting with TinyDB from a terminal. It currently has no imports from the engine packages and does not use the `server`, `plan`, or `driver` packages.

## Notes

- This package is a skeleton. `executeStatement` only prints placeholder text ("This is where we would do an insert/select.") and is NOT wired to the engine — no SQL is parsed, planned, or run.
- The only meta-command handled is `.exit`; anything else returns `META_COMMAND_UNRECOGNIZED_COMMAND`.
- Statement recognition is prefix matching for `insert`/`select` only.
- There is no exported `main`; nothing invokes `Start`. There are no test files in this package.
