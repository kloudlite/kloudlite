local ls = require("luasnip")
-- some shorthands...
local s = ls.snippet
local sn = ls.snippet_node
local t = ls.text_node
local i = ls.insert_node
local f = ls.function_node
local c = ls.choice_node
local d = ls.dynamic_node
local r = ls.restore_node
local rep = require("luasnip.extras").rep
local fmt = require("luasnip.extras.fmt").fmt
local fmta = require("luasnip.extras.fmt").fmta
local postfix = require("luasnip.extras.postfix").postfix

local snippets, autosnippets = {}, {}

function camelCase(str)
  local camelCased = ""
  local wasSeparator = false
  for i = 1, #str do
    local char = str:sub(i, i)
    if not char:match("%a") then
      wasSeparator = true
    else
      camelCased = camelCased .. (wasSeparator and char:upper() or char)
      wasSeparator = false
    end
  end
  return camelCased
end

local var = s(
  "var",
  fmta([[ {{ <p1> := get . "<p2>"}} ]], {
    p1 = f(function(...)
      local args = ...
      return "$" .. camelCase(args[1][1])
    end, 1),
    p2 = i(1, "item"),
  })
)

table.insert(snippets, var)

return snippets, autosnippets
