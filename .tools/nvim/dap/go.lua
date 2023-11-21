local dap = require("dap")

dap.configurations.go = {
	{
		type = "go",
		name = "Debug auth-api",
		request = "launch",
		program = vim.g.nxt.project_root_dir .. "/apps/auth",
		args = { "--dev" },
		console = "externalTerminal",
		-- externalTerminal = true,
		envFile = {
			vim.g.nxt.project_root_dir .. "/apps/auth" .. "/.secrets/env",
		},
	},
	{
		type = "go",
		name = "Debug infra-api",
		request = "launch",
		program = vim.g.nxt.project_root_dir .. "/apps/infra",
		args = { "--dev" },
		console = "externalTerminal",
		-- externalTerminal = true,
		envFile = {
			vim.g.nxt.project_root_dir .. "/apps/infra" .. "/.secrets/env",
		},
	},
	{
		type = "go",
		name = "Debug console-api",
		request = "launch",
		program = vim.g.nxt.project_root_dir .. "/apps/console",
		args = { "--dev" },
		console = "externalTerminal",
		-- externalTerminal = true,
		envFile = {
			vim.g.nxt.project_root_dir .. "/apps/console" .. "/.secrets/env",
		},
		-- dlvToolPath =
		-- "/usr/local/go/bin/dlv --headless=true --api-version=2 -r stdout:/tmp/debug.stdout -r stderr:/tmp/debug2.stderr",
	},
	{
		type = "go",
		name = "Debug finance-api",
		request = "launch",
		program = vim.g.nxt.project_root_dir .. "/apps/finance",
		args = { "--dev" },
		console = "externalTerminal",
		-- externalTerminal = true,
		envFile = {
			vim.g.nxt.project_root_dir .. "/apps/finance" .. "/.secrets/env",
		},
	},
	{
		type = "go",
		name = "Debug iam-api",
		request = "launch",
		program = vim.g.nxt.project_root_dir .. "/apps/iam",
		args = { "--dev" },
		console = "externalTerminal",
		-- externalTerminal = true,
		envFile = {
			vim.g.nxt.project_root_dir .. "/apps/iam" .. "/.secrets/env",
		},
	},
	{
		type = "go",
		name = "Debug message-office-api",
		request = "launch",
		program = vim.g.nxt.project_root_dir .. "/apps/message-office",
		args = { "--dev" },
		console = "externalTerminal",
		-- externalTerminal = true,
		envFile = {
			vim.g.nxt.project_root_dir .. "/apps/message-office" .. "/.secrets/env",
		},
	},
	{
		type = "go",
		name = "Debug container-registry-api",
		request = "launch",
		program = vim.g.nxt.project_root_dir .. "/apps/container-registry",
		args = { "--dev" },
		console = "externalTerminal",
		-- externalTerminal = true,
		envFile = {
			vim.g.nxt.project_root_dir .. "/apps/container-registry" .. "/.secrets/env",
		},
	},
	{
		type = "go",
		name = "Debug webhooks-api",
		request = "launch",
		program = vim.g.nxt.project_root_dir .. "/apps/webhooks",
		args = { "--dev" },
		console = "externalTerminal",
		-- externalTerminal = true,
		envFile = {
			vim.g.nxt.project_root_dir .. "/apps/webhooks" .. "/.secrets/env",
		},
	},
	{
		type = "go",
		name = "Debug accounts-api",
		request = "launch",
		program = vim.g.nxt.project_root_dir .. "/apps/accounts",
		args = { "--dev" },
		console = "externalTerminal",
		-- externalTerminal = true,
		envFile = {
			vim.g.nxt.project_root_dir .. "/apps/accounts" .. "/.secrets/env",
		},
	},
	{
		type = "go",
		name = "Debug kubelet-metrics",
		request = "launch",
		program = vim.g.nxt.project_root_dir .. "/apps/kubelet-metrics",
		args = {
			"--dev",
			"--node-name",
			"ip-172-31-13-194",
			"--enrich-from-labels",
			"--filter-prefix",
			"kloudlite.io/",
		},
		console = "externalTerminal",
		-- externalTerminal = true,
		env = {
			sample = "hello",
		},
		-- envFile = {
		--   vim.g.nxt.project_root_dir .. "/apps/kubelet-mmetrics" .. "/.secrets/env",
		-- },
	},
	{
		type = "go",
		name = "Debug comms-api",
		request = "launch",
		program = vim.g.nxt.project_root_dir .. "/apps/comms",
		args = { "--dev" },
		console = "externalTerminal",
		-- externalTerminal = true,
		envFile = {
			vim.g.nxt.project_root_dir .. "/apps/comms" .. "/.secrets/env",
		},
	},

	{
		type = "go",
		name = "Debug messages-distribution-worker",
		request = "launch",
		program = vim.g.nxt.project_root_dir .. "/apps/messages-distribution-worker",
		args = { "--dev" },
		console = "externalTerminal",
		-- externalTerminal = true,
		envFile = {
			vim.g.nxt.project_root_dir .. "/apps/messages-distribution-worker" .. "/.secrets/env",
		},
	},
}
