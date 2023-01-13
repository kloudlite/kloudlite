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

local beacon_trig = s(
  "beacon-trig",
  fmt(
    [[
	go d.beacon.TriggerWithUserCtx(ctx, {}, beacon.EventAction{{
		Action:       constants.{},
		Status:       beacon.StatusOK(),
		ResourceType: constants.{},
		ResourceId:   {},
		Tags:         map[string]string{{"projectId": {}}},
	}})
]]   ,
    {
      i(1, "/*accountId*/"),
      i(2, "/*Action*/"),
      i(3, "/*ResourceType*/"),
      i(4, "/*ResourceId*/"),
      i(5, "/*projectId*/"),
    }
  )
)

table.insert(snippets, beacon_trig)

return snippets, autosnippets
