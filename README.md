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

## Installation

```bash
go install github.com/giladkestenberg/lint-require/cmd/lint-require@latest
```

## Usage

### Standalone

```bash
lint-require ./...
```

The binary uses `singlechecker` from `go/analysis`, so it supports the standard flags:

```bash
lint-require -json ./...     # JSON output
lint-require -fix ./...      # (no auto-fixes defined yet)
lint-require ./pkg/...       # specific packages
```

### With golangci-lint (plugin)

Add to `.golangci.yml`:

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

Note: Custom module linters require building a custom golangci-lint binary. See the [golangci-lint plugin docs](https://golangci-lint.run/plugins/module-plugins/) for details.

## Neovim Integration

There are several ways to integrate `lint-require` with Neovim. Choose whichever fits your setup.

### Option 1: none-ls (null-ls successor) -- Recommended

This runs `lint-require` as a standalone diagnostic source alongside gopls.

Install [none-ls.nvim](https://github.com/nvimtools/none-ls.nvim), then add a custom source:

```lua
local null_ls = require("null-ls")

local lint_require_source = {
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
      if params.output and params.output["github.com/giladkestenberg/lint-require"] then
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
}

null_ls.setup({
  sources = {
    lint_require_source,
    -- your other sources...
  },
})
```

### Option 2: golangci-lint via none-ls

If you already use golangci-lint with the `lint-require` plugin, you get diagnostics for free through the built-in golangci-lint none-ls source:

```lua
local null_ls = require("null-ls")

null_ls.setup({
  sources = {
    null_ls.builtins.diagnostics.golangci_lint,
  },
})
```

Make sure your `.golangci.yml` has the `lintrequire` custom linter enabled (see above).

### Option 3: efm-langserver

If you use [efm-langserver](https://github.com/mattn/efm-langserver) instead of none-ls:

Add to your efm config (`~/.config/efm-langserver/config.yaml`):

```yaml
tools:
  lint-require: &lint-require
    lint-command: "lint-require -json ${INPUT}"
    lint-stdin: false
    lint-formats:
      - '%f:%l:%c: %m'

languages:
  go:
    - <<: *lint-require
```

Then configure efm in your Neovim LSP setup:

```lua
local lspconfig = require("lspconfig")

lspconfig.efm.setup({
  filetypes = { "go" },
  init_options = { documentFormatting = false },
  settings = {
    languages = {
      go = {
        { lintCommand = "lint-require ${INPUT}", lintFormats = { "%f:%l:%c: %m" } },
      },
    },
  },
})
```

### Option 4: nvim-lint

If you use [nvim-lint](https://github.com/mfussenegger/nvim-lint):

```lua
local lint = require("lint")

lint.linters.lintrequire = {
  cmd = "lint-require",
  args = { "-json" },
  stdin = false,
  stream = "stdout",
  ignore_exitcode = true,
  parser = function(output, bufnr)
    local diagnostics = {}
    local ok, decoded = pcall(vim.json.decode, output)
    if not ok then return diagnostics end
    -- Parse the singlechecker JSON output
    for _, pkg in pairs(decoded) do
      if pkg.diagnostics then
        for _, d in ipairs(pkg.diagnostics) do
          local line = d.posn and tonumber(d.posn:match(":(%d+):")) or 0
          local col = d.posn and tonumber(d.posn:match(":%d+:(%d+)")) or 0
          table.insert(diagnostics, {
            lnum = line - 1,
            col = col - 1,
            message = d.message,
            severity = vim.diagnostic.severity.ERROR,
            source = "lint-require",
          })
        end
      end
    end
    return diagnostics
  end,
}

lint.linters_by_ft = {
  go = { "lintrequire" },
}
```

Then set up an autocommand to lint on save:

```lua
vim.api.nvim_create_autocmd({ "BufWritePost" }, {
  pattern = { "*.go" },
  callback = function()
    require("lint").try_lint()
  end,
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

## Architecture

Built on Go's `go/analysis` framework. Uses `analysis.ObjectFact` to propagate required-field metadata across package boundaries, so a struct defined in package `a` with required fields will be enforced when used in package `b`.
