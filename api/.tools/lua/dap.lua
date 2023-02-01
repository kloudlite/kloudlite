local M = {}

local dap = require("dap")

M.setup = function()
  dap.configurations.go = {
    {
      type = "delve",
      request = "launch",
      program = "${file}",
      debugAdapter = "dlv",
      showLog = true,
      -- console = "externalTerminal",
      console = "internalTerminal",
      internalTerminal = true,
      externalTerminal = true,
      -- env = {
      --   hello = "hi",
      -- },
      envFile = {
        "${workspaceFolder}/.secrets/env",
        "${workspaceFolder}/.env",
      },
    },
  }

  dap.adapters.go = {
    type = "server",
    port = "${port}",
    executable = {
      command = "dlv",
      args = { "dap", "-l", "127.0.0.1:${port}" },
    },
  }

  dap.adapters.delve = {
    type = "server",
    port = "${port}",
    executable = {
      command = "dlv",
      args = { "dap", "-l", "127.0.0.1:${port}" },
    },
  }
end

return M
