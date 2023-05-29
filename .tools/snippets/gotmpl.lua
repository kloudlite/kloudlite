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
local nodeSelector = s(
  "node-selector",
  fmta(
    [[
    {{- if .Values.nodeSelector}}
    nodeSelector: {{.Values.nodeSelector | toYaml | nindent <p1>}}
    {{- end }}
    <p2>
    ]],
    {
      p1 = i(1, "4"),
      p2 = i(0),
    }
  )
)

table.insert(snippets, nodeSelector)

local tolerations = s(
  "tolerations",
  fmta(
    [[
      {{- if .Values.tolerations }}
      tolerations: {{.Values.tolerations | toYaml | nindent <p1>}}
      {{- end }}
      <p2>
    ]],
    {
      p1 = i(1, "4"),
      p2 = i(0),
    }
  )
)

table.insert(snippets, tolerations)

local imagePullPolicy = s(
  "image-pull-policy",
  fmta(
    [[
    imagePullPolicy: {{.Values.apps.<p1>.ImagePullPolicy | default .Values.imagePullPolicy }}
    <p2>
    ]],
    {
      p1 = i(1, "authApi"),
      p2 = i(0),
    }
  )
)

table.insert(snippets, imagePullPolicy)

local envKey = s(
  "env-entry",
  fmta(
    [[
key: <p1>
<p2>
]],
    {
      p1 = i(1, "//key"),
      p2 = c(2, {
        { t("value: "), i(1, "//value") },
        fmta(
          [[
          type: <np1>
          refName: <np2>
          refKey: <np3>
          ]],
          {
            np1 = c(1, { t("secret"), t("config") }),
            np2 = i(2, "config or secret name"),
            np3 = i(3, "config or secret key"),
          }
        ),
      }),
    }
  )
)

table.insert(snippets, envKey)

local metadata = s(
  "metadata",
  fmta(
    [[
    name: {{.Values.apps.<p1>.name}}
    namespace: {{.Release.Namespace}}
    labels:
      kloudlite.io/account-ref: {{.Values.accountName}}
    <p2>
]],
    {
      p1 = i(1, "authApi"),
      p2 = i(0),
    }
  )
)

table.insert(snippets, metadata)

local genValues = s("val", fmta("{{.Values.<p0>}}", { p0 = i(0) }))
table.insert(snippets, genValues)

local if_stmt = s(
  "if",
  fmta(
    [[
{{- if <p1> }}
<p0>
{{- end }}
]],
    {
      p1 = i(1, "condition"),
      p0 = i(0),
    }
  )
)

table.insert(snippets, if_stmt)

local expandToValues = s(".v", t(".Values."))
local expandToValues2 = s("{{.v", t("{{.Values."))

table.insert(autosnippets, expandToValues)
table.insert(autosnippets, expandToValues2)

return snippets, autosnippets
