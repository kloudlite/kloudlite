local dap = require("dap")

dap.configurations.go = {
  {
    type = "go",
    name = "Debug infra-api",
    request = "launch",
    program = vim.g.root_dir .. "/apps/infra",
    args = { "--dev" },
    console = "externalTerminal",
    -- externalTerminal = true,
    envFile = {
      vim.g.root_dir .. "/apps/infra" .. "/.secrets/env",
    },
  },
}
