local dap = require("dap")

dap.configurations.go = {
  {
    type = "go",
    name = "Debug app-n-lambda",
    request = "launch",
    program = vim.g.root_dir .. "/operators/app-n-lambda",
    args = { "--dev" },
    console = "externalTerminal",
    -- externalTerminal = true,
    envFile = {
      vim.g.root_dir .. "/operators/app-n-lambda" .. "/.secrets/env",
    },
  },
  {
    type = "go",
    name = "Debug artifacts-harbor",
    request = "launch",
    program = vim.g.root_dir .. "/operators/artifacts-harbor",
    args = { "--dev" },
    console = "externalTerminal",
    -- externalTerminal = true,
    envFile = {
      vim.g.root_dir .. "/operators/artifacts-harbor" .. "/.secrets/env",
    },
  },
  {
    type = "go",
    name = "Debug routers",
    request = "launch",
    program = vim.g.root_dir .. "/operators/routers",
    args = { "--dev" },
    console = "externalTerminal",
    -- externalTerminal = true,
    envFile = {
      vim.g.root_dir .. "/operators/routers" .. "/.secrets/env",
    },
  },
  {
    type = "go",
    name = "Debug cluster-setup",
    request = "launch",
    program = vim.g.root_dir .. "/operators/cluster-setup",
    args = { "--dev" },
    -- console = "externalTerminal",
    -- externalTerminal = true,
    envFile = {
      vim.g.root_dir .. "/operators/cluster-setup" .. "/.secrets/env",
    },
  },
  {
    type = "go",
    name = "Debug status-n-billing",
    request = "launch",
    program = vim.g.root_dir .. "/operators/status-n-billing",
    args = { "--dev", "--serverHost", "localhost:8081" },
    -- console = "externalTerminal",
    -- externalTerminal = true,
    envFile = {
      vim.g.root_dir .. "/operators/status-n-billing" .. "/.secrets/env",
    },
  },
  {
    type = "go",
    name = "Debug agent",
    request = "launch",
    program = vim.g.root_dir .. "/agent",
    args = { "--dev" },
    -- console = "externalTerminal",
    -- externalTerminal = true,
    envFile = {
      vim.g.root_dir .. "/agent" .. "/.secrets/env",
    },
  },
  {
    type = "go",
    name = "Debug byoc-helm-status-watcher",
    request = "launch",
    program = vim.g.root_dir .. "/operators/byoc-helm-status-watcher",
    args = { "--dev" },
    console = "externalTerminal",
    -- externalTerminal = true,
    envFile = {
      vim.g.root_dir .. "/operators/byoc-helm-status-watcher" .. "/.secrets/env",
    },
  },
  {
    type = "go",
    name = "Debug byoc-operator",
    request = "launch",
    program = vim.g.root_dir .. "/operators/byoc-operator",
    args = { "--dev" },
    console = "externalTerminal",
    -- externalTerminal = true,
    envFile = {
      vim.g.root_dir .. "/operators/byoc-operator" .. "/.secrets/env",
    },
  },
  {
    type = "go",
    name = "Debug byoc-client-operator",
    request = "launch",
    program = vim.g.root_dir .. "/operators/byoc-client-operator",
    args = { "--dev" },
    console = "externalTerminal",
    -- externalTerminal = true,
    envFile = {
      vim.g.root_dir .. "/operators/byoc-client-operator" .. "/.secrets/env",
    },
  },
  {
    type = "go_test",
    name = "[Debug] Test app-n-lambda",
    request = "remote",
    mode = "test",
    program = vim.g.root_dir .. "/operators/app-n-lambda/internal/controllers/app/control",
    env = {
      PROJECT_ROOT = vim.g.root_dir,
    },
  },
  {
    type = "go",
    name = "Debug msvc-mongo",
    request = "launch",
    program = vim.g.root_dir .. "/operators/msvc-mongo",
    args = { "--dev" },
    console = "externalTerminal",
    -- externalTerminal = true,
    envFile = {
      vim.g.root_dir .. "/operators/msvc-mongo" .. "/.secrets/env",
    },
  },
}
