# lint-require

A Go linter that enforces required struct fields. Annotate struct fields with `// required: <reason>` comments, and the linter will report an error whenever a composite literal of that struct omits a required field.

Works across package boundaries -- define the struct in one package, use it in another, and the linter still catches missing fields.

## How it works

Annotate struct fields with a `// required:` comment:

```go
type Config struct {
    // required: name must be set for identification
    Name string
    // required: port is needed for networking
    Port     int
    Optional string
}
```

The linter flags any composite literal missing a required field:

```go
_ = Config{Name: "app"}           // ERROR: missing required field "Port": port is needed for networking
_ = Config{Name: "app", Port: 80} // OK
_ = Config{}                       // ERROR: missing both "Name" and "Port"
```

Positional literals (e.g. `Config{"app", 80, ""}`) are not checked since all fields are provided.

## Install

### 1. Install the binary

```bash
go install github.com/GKestenberg/lint-require/cmd/lint-require@latest
```

### 2. Add to golangci-lint

Add to your `.golangci.yml`:

```yaml
linters-settings:
  custom:
    lintrequire:
      type: module
      description: "Checks that struct fields with // required: comments are present in composite literals"

linters:
  enable:
    - lintrequire
```

> **Note:** Custom module linters require building a custom golangci-lint binary. See the [golangci-lint plugin docs](https://golangci-lint.run/plugins/module-plugins/) for details.

### 3. Neovim setup (golangci-lint + none-ls)

Install [none-ls.nvim](https://github.com/nvimtools/none-ls.nvim) if you haven't already, then add the golangci-lint diagnostics source. This picks up `lint-require` automatically through your `.golangci.yml`:

```lua
local null_ls = require("null-ls")

null_ls.setup({
  sources = {
    null_ls.builtins.diagnostics.golangci_lint,
  },
})
```

That's it -- diagnostics show inline in your editor on save.

## Standalone usage

You can also run it directly without golangci-lint:

```bash
lint-require ./...
lint-require -json ./...     # JSON output
lint-require ./pkg/...       # specific packages
```

### Standalone with none-ls (without golangci-lint)

If you prefer to run `lint-require` directly instead of through golangci-lint:

```lua
local null_ls = require("null-ls")

null_ls.setup({
  sources = {
    {
      name = "lint-require",
      method = null_ls.methods.DIAGNOSTICS_ON_SAVE,
      filetypes = { "go" },
      generator = null_ls.generator({
        command = "lint-require",
        args = { "-json", "$DIRNAME" },
        to_stdin = false,
        format = "json",
        check_exit_code = function(code)
          return code <= 2
        end,
        on_output = function(params)
          local diagnostics = {}
          if params.output then
            for _, pkg in pairs(params.output) do
              if pkg.diagnostics then
                for _, d in ipairs(pkg.diagnostics) do
                  table.insert(diagnostics, {
                    row = d.posn and tonumber(d.posn:match(":(%d+):")) or 1,
                    col = d.posn and tonumber(d.posn:match(":%d+:(%d+)")) or 1,
                    message = d.message,
                    severity = vim.diagnostic.severity.ERROR,
                    source = "lint-require",
                  })
                end
              end
            end
          end
          return diagnostics
        end,
      }),
    },
  },
})
```

## Comment Syntax

```
// required: <reason>
```

- Must be a doc comment directly above the struct field (no blank line between)
- Case-insensitive match on `required:`
- Everything after the colon is used as the error message reason
- If no reason is provided, defaults to "field is required"
