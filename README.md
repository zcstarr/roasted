# Roasted

This project is a Go library to interface with the Freshroast SR-700 coffee roaster. This is a fork of [jmhobbs/roasted](https://github.com/jmhobbs/roasted).

All the heavy lifting reversing the protocol was done by the excellent [Openroast](https://github.com/Roastero/Openroast) project, this is just an implementation of that protocol.

## Recipes

Recipes are JSON files. Two formats are supported, both described by JSON Schemas in [`schemas/`](./schemas):

- `SimpleRecipe` (`*.recipe.json`) — what the `roasted` CLI runs. Each step has `heat` (0-3), `fan` (1-9), `duration` (seconds), and optional `cooling`.
- `OpenRoastRecipe` (`*.openroast.json`) — the Openroast project's format, with `fanSpeed`, `sectionTime`, `targetTemp`, and bean metadata.

VS Code / Cursor pick up the schemas automatically (see [`.vscode/settings.json`](./.vscode/settings.json)), so you get autocomplete, hover docs, and validation while editing. A short example lives at [`sr700-fan-test.recipe.json`](./sr700-fan-test.recipe.json).

The schemas are generated from the Go types in [`pkg/recipe.go`](./pkg/recipe.go). Edit the types, then run `make generate` and commit the result.

## Development

| Command | Purpose |
| --- | --- |
| `make help` | List available targets |
| `make build` | Compile the library and CLI |
| `make install` | Install the `roasted` CLI |
| `make test` | Run unit tests with race detector |
| `make generate` | Regenerate JSON schemas from Go types |
| `make lint` | Run `golangci-lint` |
| `make check` | Lint + test + schema-drift check (the CI gate) |

Run `make check` before pushing.
