# Providers & families — how identification works without lookup tables

## Providers

Evidence is collected per model:

| Field | Source |
|---|---|
| `owned_by` | `GET /v1/models` |
| id prefix | the part before the first `/` of the model id |
| connection id | gateway error wrappers such as `[groq/llama-3.3-70b] [429]: …` |
| served model | `model` field of successful responses |

For each candidate field the resolver computes its **discriminating power** (number of distinct values and their
distribution over the model set) and its **consistency** with the connection ids observed in errors. The field that
distinguishes providers *and* agrees with connection ids becomes the key. If `owned_by` is constant for every model
(e.g. always `openai` or `library`) the id prefix is used; if that does not distinguish either, the endpoint is treated
as a single provider named after itself. The rationale ("owned_by distinguishes 40 values and matches 1 877
connection ids") is shown in the dashboard and stored in exports.

Display names are the most readable raw value of the key (never a translated table). Users can rename or merge
providers per endpoint; the resolver also emits **merge suggestions** when two keys share connection ids or error
domains. Suggestions are never applied silently.

## Families

A family groups the variants of one model line (`deepseek-v4-flash`, `deepseek-v4.1-flash-0731`,
`DeepSeek-V4-Flash:free`, `dsv41flash`, …). Grouping has three layers:

1. **Tokeniser rules** (language-level, endpoint-agnostic): strip provider/organisation prefixes and `:free`-style
   suffixes, parameter sizes (`7b`, `a3b`, `8x22b`), dates, quantisation marks and modifiers (`preview`,
   `instruct`, `thinking`, `free`, `fast`, …); fold major versions (`v4.1 → 4`, `5.3 → 5`); keep tier/speciality words
   (`flash`, `pro`, `max`, `mini`, `nano`, `vl`, `coder`, …).
2. **Alias learning** from data: a model that *serves* another name (request `deepseek-chat`, response
   `deepseek-v4-flash-0731`) joins the served family; short ids are expanded against the vocabulary of the endpoint's
   own model list (`dsv41flash → deepseek|flash|4`) only when the match is unique; equal context sizes act as weak
   corroboration. Generic pseudo-models (`auto`, `free`, routers) are flagged and never absorb other families.
3. **Overrides** per endpoint (`family_overrides`): assign a model to a family key and optional canonical display name.

Every assignment records its *source* (`rule`, `served`, `abbrev`, `override`) and evidence; the family drawer shows the
mix of sources so you can judge how much of a group is inferred.

## Managing overrides

- Web UI: *Endpoints → Providers & families* dialog, or the *Providers* tab of a run (rename / merge).
- HTTP API: `PUT /api/endpoints/{id}/overrides/provider` `{raw_key, display_name, merged_into}` and
  `PUT /api/endpoints/{id}/overrides/family` `{model_id, family_key, canon}`; `GET /api/endpoints/{id}/overrides`.
- MCP: `get_providers` exposes evidence and suggestions; overrides are applied through the API.

Clearing the display name and merge target restores automatic identification for that key.
