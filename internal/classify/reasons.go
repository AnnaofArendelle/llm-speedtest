package classify

// ReasonInfo documents a reason code for reports and the UI.
type ReasonInfo struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	NameEN        string `json:"name_en"`
	Description   string `json:"description"`
	DescriptionEN string `json:"description_en"`
	Fix           string `json:"fix"`
	FixEN         string `json:"fix_en"`
	ProviderLevel bool   `json:"provider_level"`
	Transient     bool   `json:"transient"`
}

// ReasonSkipped marks attempts cancelled by the user through the skip feature.
const ReasonSkipped = "skipped"

// Order lists reason codes in presentation order.
var Order = []string{
	"no_credentials", "auth", "quota_billing", "quota_exhausted", "plan_restriction", "terms_required", "rate_limit",
	"model_unavailable", "model_retired", "not_chat_model", "not_chat_model_inferred", "param_incompatible", "no_capacity",
	"no_channel", "method_not_allowed", "empty_response", "upstream_network", "upstream_5xx", "timeout", "network",
	"bad_request", "error_in_200", "exception", "unknown", "skipped", "untested",
}

var providerLevel = map[string]bool{
	"no_credentials": true, "auth": true, "quota_billing": true, "quota_exhausted": true, "plan_restriction": true, "terms_required": true,
}

var transient = map[string]bool{
	"rate_limit": true, "upstream_5xx": true, "upstream_network": true, "timeout": true, "network": true, "empty_response": true,
	"no_capacity": true, "no_channel": true, "exception": true, "unknown": true, "error_in_200": true,
}

var reasons = map[string]ReasonInfo{
	"ok": {Code: "ok", Name: "可用", NameEN: "Available"},
	"no_credentials": {Code: "no_credentials", Name: "未配置凭证", NameEN: "No credentials configured",
		Description:   "网关中该提供商没有激活的凭证，请求在网关本地即被拒绝（HTTP 404），没有发往上游。",
		DescriptionEN: "The gateway has no active credential for this provider; the request was rejected locally (HTTP 404) and never reached the upstream.",
		Fix:           "在网关中为该提供商添加或启用凭证。", FixEN: "Add or enable a credential for the provider in the gateway."},
	"auth": {Code: "auth", Name: "认证失败 (401/403)", NameEN: "Authentication failed (401/403)",
		Description:   "上游拒绝了凭证：OAuth token 失效/过期、API Key 无效或被停用。整个提供商下的模型都会同样失败。",
		DescriptionEN: "The upstream rejected the credential: expired OAuth token, invalid or disabled API key. Every model behind the provider fails the same way.",
		Fix:           "重新登录 OAuth 或更换 API Key。", FixEN: "Re-authenticate or replace the API key."},
	"quota_billing": {Code: "quota_billing", Name: "余额不足，需充值", NameEN: "Insufficient balance",
		Description:   "上游账号余额为 0、积分不足或预扣费失败，拒绝了计费请求。属于账号级问题，与具体模型无关。",
		DescriptionEN: "The upstream account has no balance or credits left; billing pre-authorisation failed. Account-level, unrelated to the model.",
		Fix:           "充值；免费额度类账号可更换账号。", FixEN: "Top up the account or switch accounts."},
	"quota_exhausted": {Code: "quota_exhausted", Name: "免费额度已用尽", NameEN: "Free quota exhausted",
		Description:   "报错表明账号级的免费请求额度 / 免费模型配额 / 赠送 token 已用完（通常为 429），需等待窗口重置，而非瞬时限流。",
		DescriptionEN: "The free request quota, free-model allowance or gifted tokens for the account are used up (usually 429); it resets with the quota window rather than being a momentary rate limit.",
		Fix:           "等待额度窗口重置、充值或更换账号。", FixEN: "Wait for the quota window to reset, top up or switch accounts."},
	"plan_restriction": {Code: "plan_restriction", Name: "套餐限制", NameEN: "Plan restriction",
		Description:   "当前免费套餐不允许使用该（付费/高级）模型，或平台已暂停免费套餐，或账号无会员资格。",
		DescriptionEN: "The current (free) plan does not allow this paid/premium model, the free plan is paused, or the account lacks a membership.",
		Fix:           "升级套餐，或只使用该平台标记为 free 的模型。", FixEN: "Upgrade the plan or use the models marked as free."},
	"terms_required": {Code: "terms_required", Name: "需先接受模型条款", NameEN: "Terms acceptance required",
		Description:   "提供商要求组织管理员先在控制台接受该模型的使用条款。",
		DescriptionEN: "The provider requires an organisation admin to accept the model's terms of use in the console first.",
		Fix:           "到提供商控制台接受条款后即可使用。", FixEN: "Accept the terms in the provider console."},
	"rate_limit": {Code: "rate_limit", Name: "限流 (429)", NameEN: "Rate limited (429)",
		Description:   "上游瞬时 RPM/TPM 限制、按分钟计数的免费层请求上限，或聚合平台“所有渠道繁忙”；放慢并复测后仍失败。",
		DescriptionEN: "Momentary RPM/TPM limits, per-minute free-tier caps or 'all channels busy' on an aggregator; still failing after slower retries.",
		Fix:           "降低调用频率、稍后重试。", FixEN: "Reduce the request rate and retry later."},
	"model_unavailable": {Code: "model_unavailable", Name: "模型不存在/不可用 (404/400)", NameEN: "Model not found / unavailable (404/400)",
		Description:   "上游不认识该模型名、账号无权访问、模型已停用或暂时禁用；也包含网关本地直接判定“不支持”的模型。属于模型级问题。",
		DescriptionEN: "The upstream does not know the model, the account cannot access it, or it is disabled; includes models the gateway rejects locally. Model-level.",
		Fix:           "从列表移除，或改用上游当前有效的模型名。", FixEN: "Remove it from the list or switch to a currently valid model name."},
	"model_retired": {Code: "model_retired", Name: "模型已下线 (410)", NameEN: "Model retired (410)",
		Description:   "上游明确告知该模型已停用/到期，多数给出了替代模型。",
		DescriptionEN: "The upstream states the model reached end of life; most responses name a replacement.",
		Fix:           "按报错提示改用替代模型。", FixEN: "Use the replacement model named in the error."},
	"not_chat_model": {Code: "not_chat_model", Name: "非聊天模型", NameEN: "Not a chat model",
		Description:   "上游明确说明该模型不支持 /v1/chat/completions（图像生成、OCR、语音、决策等专用端点）。",
		DescriptionEN: "The upstream states the model does not support /v1/chat/completions (image generation, OCR, speech or other dedicated endpoints).",
		Fix:           "改用对应的专用端点。", FixEN: "Use the dedicated endpoint for that modality."},
	"not_chat_model_inferred": {Code: "not_chat_model_inferred", Name: "非聊天模型（按模型名推断）", NameEN: "Not a chat model (inferred from name)",
		Description:   "上游报“模型不存在或不支持 chat completions”，而模型名表明它是图像/视频/语音/嵌入类模型，本来就不走 chat 接口。",
		DescriptionEN: "The upstream reported 'not found or not supported for chat', and the name indicates an image/video/audio/embedding model that never used the chat endpoint.",
		Fix:           "改用对应的专用端点。", FixEN: "Use the dedicated endpoint for that modality."},
	"param_incompatible": {Code: "param_incompatible", Name: "请求参数不兼容", NameEN: "Incompatible request parameters",
		Description:   "模型存在，但上游拒绝了网关发出的请求参数（如 thinking level、输入类型）。",
		DescriptionEN: "The model exists but the upstream rejected a request parameter forwarded by the gateway (thinking level, input type, ...).",
		Fix:           "调整该提供商/模型的参数策略后重试。", FixEN: "Adjust the parameter policy for the provider/model and retry."},
	"no_capacity": {Code: "no_capacity", Name: "暂无可用容量", NameEN: "No capacity available",
		Description:   "聚合平台回复“当前没有可用容量，请稍后再试”；最终复测仍失败。",
		DescriptionEN: "The aggregator replied 'no capacity available right now'; still failing on the final retry.",
		Fix:           "稍后重试。", FixEN: "Retry later."},
	"no_channel": {Code: "no_channel", Name: "平台无可用渠道", NameEN: "No channel on the platform",
		Description:   "聚合平台报该模型在当前分组下没有可用渠道/部署，属于该平台的配置问题。",
		DescriptionEN: "The aggregator has no channel/deployment for the model in the current group; a platform configuration issue.",
		Fix:           "换用其它提供商的同名模型。", FixEN: "Use the same model from a different provider."},
	"method_not_allowed": {Code: "method_not_allowed", Name: "接口不支持 (405)", NameEN: "Endpoint not supported (405)",
		Description:   "网关的该提供商适配器不支持 chat completions 调用方式。",
		DescriptionEN: "The gateway adapter for this provider does not support the chat completions call.",
		Fix:           "改用该提供商原生接口，或等待网关适配。", FixEN: "Use the provider's native API or wait for gateway support."},
	"empty_response": {Code: "empty_response", Name: "上游返回空内容", NameEN: "Empty upstream response",
		Description:   "上游连接成功但没有返回任何内容（复测仍如此）。",
		DescriptionEN: "The upstream connection succeeded but returned no content (also on retry).",
		Fix:           "更换模型或稍后重试。", FixEN: "Switch models or retry later."},
	"upstream_network": {Code: "upstream_network", Name: "上游网络故障", NameEN: "Upstream network failure",
		Description:   "网关连接上游时 fetch failed / 连接被重置 / TLS 握手失败 / 网关错误。",
		DescriptionEN: "The gateway failed to reach the upstream: fetch failed, connection reset, TLS handshake failure or bad gateway.",
		Fix:           "检查代理与上游域名可达性。", FixEN: "Check proxies and upstream reachability."},
	"upstream_5xx": {Code: "upstream_5xx", Name: "上游 5xx", NameEN: "Upstream 5xx",
		Description:   "上游服务端内部错误或暂时不可用；复测仍失败。",
		DescriptionEN: "Upstream server error or temporary unavailability; still failing on retry.",
		Fix:           "稍后重试。", FixEN: "Retry later."},
	"skipped": {Code: "skipped", Name: "用户跳过", NameEN: "Skipped by user",
		Description:   "用户在运行过程中跳过了该模型：在途请求被取消，后续轮次与阶段不再测试它。",
		DescriptionEN: "The user skipped this model while the run was in progress: the in-flight request was cancelled and later rounds and stages leave it out.",
		Fix:           "重新运行以获得该模型的结果；通常意味着它响应过慢。", FixEN: "Re-run to measure it; usually it means the model was too slow to respond."},
	"timeout": {Code: "timeout", Name: "超时", NameEN: "Timeout",
		Description:   "客户端在超时阈值内没有收到完整响应。",
		DescriptionEN: "No complete response arrived within the client timeout.",
		Fix:           "模型过慢或请求挂起；可单独用更长超时验证。", FixEN: "The model is slow or the request hung; verify with a longer timeout."},
	"network": {Code: "network", Name: "本地网络错误", NameEN: "Local network error",
		Description: "无法连接端点。", DescriptionEN: "The endpoint could not be reached.",
		Fix: "检查端点地址、代理与网络。", FixEN: "Check the endpoint URL, proxy and network."},
	"bad_request": {Code: "bad_request", Name: "请求被拒绝 (400)", NameEN: "Request rejected (400)",
		Description:   "上游或网关以 400 拒绝请求（输入校验失败等，非“模型不存在”类）。",
		DescriptionEN: "The upstream or gateway rejected the request with 400 (validation failure, not a missing model).",
		Fix:           "查看报错原文。", FixEN: "Inspect the error message."},
	"error_in_200": {Code: "error_in_200", Name: "200 但含错误体", NameEN: "HTTP 200 with error body",
		Description: "上游返回 HTTP 200，但响应体是错误对象。", DescriptionEN: "The upstream returned HTTP 200 but the body is an error object.",
		Fix: "查看报错原文。", FixEN: "Inspect the error message."},
	"exception": {Code: "exception", Name: "客户端异常", NameEN: "Client exception",
		Description: "测试客户端内部异常。", DescriptionEN: "Internal client error.", Fix: "查看报错原文。", FixEN: "Inspect the error message."},
	"unknown": {Code: "unknown", Name: "未知错误", NameEN: "Unknown error",
		Description: "无法归类。", DescriptionEN: "Could not be classified.", Fix: "查看报错原文。", FixEN: "Inspect the error message."},
	"untested": {Code: "untested", Name: "未测试", NameEN: "Not tested",
		Description: "没有任何测试记录。", DescriptionEN: "No attempts recorded."},
}

// Info returns the documentation for a reason code; unknown codes yield a
// minimal entry using the code as its name.
func Info(code string) ReasonInfo {
	if r, ok := reasons[code]; ok {
		r.ProviderLevel = providerLevel[code]
		r.Transient = transient[code]
		return r
	}
	return ReasonInfo{Code: code, Name: code, NameEN: code}
}

// All returns every documented reason in presentation order.
func All() []ReasonInfo {
	out := make([]ReasonInfo, 0, len(Order))
	for _, c := range Order {
		out = append(out, Info(c))
	}
	return out
}
