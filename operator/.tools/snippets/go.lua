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
local extras = require("luasnip.extras")
local l = extras.lambda
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
]],
    {}
  )
)

table.insert(snippets, res_enabled)

local robj = s(
  "robj",
  fmta(
    [[
	ctx, obj := req.Context(), req.Object
	check := rApi.Check{Generation: obj.Generation}

	req.LogPreCheck(<p1>)
	defer req.LogPostCheck(<p2>)

  <p3>

	check.Status = true
	if check != obj.Status.Checks[<p4>] {
		obj.Status.Checks[<p5>] = check
		req.UpdateStatus()
	}

	return req.Next()
]],
    {
      p1 = i(1, "Checkname"),
      p2 = rep(1),
      p3 = i(0, "//body"),
      p4 = rep(1),
      p5 = rep(1),
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
]],
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
]],
    {}
  )
)
table.insert(snippets, import_ginkgo)

local commonTypesImports = s(
  "imp_k8s_types",
  fmt(
    [[
	rApi "github.com/kloudlite/operator/pkg/operator"
	stepResult "github.com/kloudlite/operator/pkg/operator/step-result"
	fn "github.com/kloudlite/operator/pkg/functions"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrl "sigs.k8s.io/controller-runtime"
]],
    {}
  )
)

table.insert(snippets, commonTypesImports)

-- local kubebuilder_marker = s("k_marker", fmta("// +kubebuilder:<p1>=<p2>", {
-- }))
-- local rr = postfix(
--   { trig = ".rr", match_pattern = ".*" },
--   fmta(
--     [[if <err> != nil {
--         return <result>
--       }
--      <finish>
--     ]],
--     {
--       err = i(1, "err"),
--       result = d(2, go_ret_vals, { 1, 2 }),
--       finish = i(0),
--     }
--   )
-- )

-- local stexpr = postfix({ trig = ".stexpr" },
--   fmta([[
-- if step := <expr>; !step.ShouldProceed() {
--   return step.ReconcilerResponse()
-- }
-- <finish>
-- ]], {
--     expr = f(1, function(_, parent)
--       return parent.env.POSTFIX_MATCH
--     end),
--     finish = i(0),
--   })
-- )

local stexpr = postfix(".stexpr", {
  f(function(_, parent)
    return "[" .. parent.snippet.env.POSTFIX_MATCH .. "]"
  end, {}),
})
table.insert(snippets, stexpr)

return snippets, autosnippets
