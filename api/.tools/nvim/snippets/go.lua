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
]],
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

local imp_crdsv1 = s("imp_crdsv1", t('crdsv1 "github.com/kloudlite/operator/apis/crds/v1"'))
table.insert(snippets, imp_crdsv1)

local gql_marshaler = s(
  "gql_marshaler",
  fmta(
    [[
func (<p1> *<p2>) UnmarshalGQL(v interface{}) error {
  switch t := v.(type) {
    case map[string]any:
      b, err := json.Marshal(t)
      if err != nil {
        return err
      }

      if err := json.Unmarshal(b, <p3>); err != nil {
        return err
      }

    case string:
      if err := json.Unmarshal([]byte(t), <p4>); err != nil {
        return err
      }
  }

	return nil
}

func (<p5> <p6>) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(<p7>)
	if err != nil {
		w.Write([]byte("{}"))
	}
	w.Write(b)
}
]],
    {
      p1 = i(1, "obj"),
      p2 = i(2, "//type"),
      p3 = rep(1),
      p4 = rep(1),
      p5 = rep(1),
      p6 = rep(2),
      p7 = rep(1),
    }
  )
)
table.insert(snippets, gql_marshaler)

local fx_lifecycle_hook = s(
  "fx_lifecycle_hook",
  fmta(
    [[
	fx.Invoke(func(lf fx.Lifecycle) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
			  return <on_start>
			},
			OnStop: func(ctx context.Context) error {
			  return <on_stop>
			},
		})
	}),
  ]],
    {
      on_start = i(1, "nil"),
      on_stop = i(2, "nil"),
    }
  )
)

table.insert(snippets, fx_lifecycle_hook)

return snippets, autosnippets
