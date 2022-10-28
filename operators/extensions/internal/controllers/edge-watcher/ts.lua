local function P(item)
  print(vim.inspect(item))
end

local bufnr = 25

local langTree = vim.treesitter.get_parser(bufnr, "go")
-- P(langTree:lang())

local syntaxTree = langTree:parse()
-- P(getmetatable(syntaxTree[1]))

local root = syntaxTree[1]:root()

for c in root:iter_children() do
  P(c:type())
end

-- local query = vim.treesitter.parse_query(
--   "go",
--   [[
--     (method_declaration) @method
--   ]]
-- )
--
-- for _, captures, metadata in query:iter_matches(root, bufnr) do
--   P(captures[1]:start())
-- end
