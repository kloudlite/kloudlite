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

local res_enabled = s(
  "res_enabled",
  fmt(
    [[
// +kubebuilder:default=true
Enabled bool `json:"enabled,omitempty"`
]]   ,
    {}
  )
)

table.insert(snippets, res_enabled)

local robj = s(
  "robj",
  fmt(
    [[
	ctx, obj, checks := req.Context(), req.Object, req.Object.Status.Checks
	check := rApi.Check{{Generation: obj.Generation}}

	req.LogPreCheck({})
	defer req.LogPostCheck({})

	check.Status = true
	if check != checks[{}] {{
		checks[{}] = check
		return req.UpdateStatus()
	}}

	return req.Next()
]]   ,
    {
      i(1, "Checkname"),
      rep(1),
      rep(1),
      rep(1),
    }
  )
)
table.insert(snippets, robj)

local import_ginkgo = s(
  "imp_ginkgo",
  fmt(
    [[
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
]]   ,
    {}
  )
)
table.insert(snippets, import_ginkgo)

local import_test_suite = s(
  "imp_suite",
  fmt(
    [[
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
]]   ,
    {}
  )
)
table.insert(snippets, import_ginkgo)

return snippets, autosnippets
