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
		args = { "--dev" },
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
		type = "go_test",
		name = "[Debug] Test app-n-lambda",
		request = "remote",
		mode = "test",
		program = vim.g.root_dir .. "/operators/app-n-lambda/internal/controllers/app/control",
		env = {
			PROJECT_ROOT = vim.g.root_dir,
		},
	},
}
