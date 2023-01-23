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

local imp_crdsv1 = s("imp_crdsv1", t('crdsv1 "github.com/kloudlite/operator/apis/crds/v1"'))
table.insert(snippets, imp_crdsv1)

local gql_marshaler = s(
  "gql_marshaler",
  fmta(
    [[
func (<> *<>) UnmarshalGQL(v interface{}) error {
	if err := json.Unmarshal([]byte(v.(string)), er); err != nil {
		return err
	}

	// if err := validator.Validate(*c); err != nil {
	// 	return err
	// }

	return nil
}

func (<> <>) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(er)
	if err != nil {
		w.Write([]byte("{}"))
	}
	w.Write(b)
}
]]   ,
    {
      i(1, "obj"),
      i(2, "//type"),
      rep(1),
      rep(2),
    }
  )
)
table.insert(snippets, gql_marshaler)

return snippets, autosnippets
