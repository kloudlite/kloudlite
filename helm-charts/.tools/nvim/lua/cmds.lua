local utils = require("nxtcoder17.utils.strings")

vim.api.nvim_create_user_command("Helm", function()
	vim.cmd("vsplit")
	vim.cmd("wincmd l")
	local handle = io.popen("cat Chart.yaml | yq -r '.name'")
	local result = handle:read("*a")
	handle:close()
	local renderedPath = string.format(
		"/tmp/manifests/%s/templates/%s",
		utils.trim(result),
		utils.trim(vim.fn.expand("%:p:t"))
	)
	vim.cmd("e " .. renderedPath)
	-- vim.cmd("set ft=gotmpl | setlocal buftype=nofile | setlocal bufhidden=hide | setlocal noswapfile")
end, { force = true })
