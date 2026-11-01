import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, encodePath } from './client';
import type {
  Attempt,
  CapMatrix,
  CheckResult,
  CompareDoc,
  DiscoverResult,
  Endpoint,
  FailureGroup,
  FamilyStat,
  LatencyModelStats,
  LatencySample,
  LoadLevel,
  ModelStat,
  ProbeInfo,
  ProviderStat,
  QualityScore,
  ReasonInfo,
  Run,
  RunEvent,
  RunSummaryDoc,
  SuiteSpec,
} from './types';

export const keys = {
  version: ['version'] as const,
  reasons: ['reasons'] as const,
  probes: ['probes'] as const,
  presets: ['presets'] as const,
  dialects: ['dialects'] as const,
  endpoints: ['endpoints'] as const,
  overrides: (id: string) => ['overrides', id] as const,
  runs: (kind?: string) => ['runs', kind ?? ''] as const,
  run: (id: string) => ['run', id] as const,
  summary: (id: string) => ['run', id, 'summary'] as const,
  families: (id: string, order: string) => ['run', id, 'families', order] as const,
  providers: (id: string) => ['run', id, 'providers'] as const,
  models: (id: string, q: string) => ['run', id, 'models', q] as const,
  model: (id: string, m: string) => ['run', id, 'model', m] as const,
  failures: (id: string) => ['run', id, 'failures'] as const,
  latency: (id: string) => ['run', id, 'latency'] as const,
  load: (id: string) => ['run', id, 'load'] as const,
  capabilities: (id: string) => ['run', id, 'capabilities'] as const,
  quality: (id: string) => ['run', id, 'quality'] as const,
  compare: (a: string, b: string) => ['compare', a, b] as const,
};

export function useVersion() {
  return useQuery({
    queryKey: keys.version,
    queryFn: () => api<{ version: string; commit: string; date: string; go_version: string }>('/api/version'),
    staleTime: Infinity,
  });
}
export function useReasons() {
  return useQuery({ queryKey: keys.reasons, queryFn: () => api<ReasonInfo[]>('/api/reasons'), staleTime: Infinity });
}
export function useProbes() {
  return useQuery({ queryKey: keys.probes, queryFn: () => api<ProbeInfo[]>('/api/probes'), staleTime: Infinity });
}
export function usePresets() {
  return useQuery({
    queryKey: keys.presets,
    queryFn: () => api<Record<string, SuiteSpec>>('/api/presets'),
    staleTime: Infinity,
  });
}
export function useDialects() {
  return useQuery({ queryKey: keys.dialects, queryFn: () => api<string[]>('/api/dialects'), staleTime: Infinity });
}

export function useEndpoints() {
  return useQuery({ queryKey: keys.endpoints, queryFn: () => api<Endpoint[]>('/api/endpoints') });
}

export interface EndpointInput {
  id?: string;
  name: string;
  base_url: string;
  api_key?: string;
  clear_key?: boolean;
  dialect?: string;
  headers?: Record<string, string>;
  insecure?: boolean;
  proxy?: string;
}

export function useSaveEndpoint() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: EndpointInput) =>
      input.id
        ? api<Endpoint>(`/api/endpoints/${input.id}`, { method: 'PUT', json: input })
        : api<Endpoint>('/api/endpoints', { method: 'POST', json: input }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.endpoints }),
  });
}

export function useDeleteEndpoint() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api(`/api/endpoints/${id}`, { method: 'DELETE' }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.endpoints }),
  });
}

export function useCheckEndpoint() {
  return useMutation({
    mutationFn: (id: string) => api<CheckResult>(`/api/endpoints/${id}/check`, { method: 'POST' }),
  });
}

export function useCheckAdhoc() {
  return useMutation({
    mutationFn: (input: { base_url: string; api_key?: string; dialect?: string }) =>
      api<CheckResult>('/api/check', {
        method: 'POST',
        json: { endpoint: { base_url: input.base_url, dialect: input.dialect ?? '' }, api_key: input.api_key ?? '' },
      }),
  });
}

export function useDiscoverAdhoc() {
  return useMutation({
    mutationFn: (input: { base_url: string; api_key?: string; dialect?: string; filter?: unknown }) =>
      api<DiscoverResult>('/api/discover', {
        method: 'POST',
        json: {
          endpoint: { base_url: input.base_url, dialect: input.dialect ?? '' },
          api_key: input.api_key ?? '',
          filter: input.filter,
        },
      }),
  });
}

export function useDiscoverEndpoint() {
  return useMutation({
    mutationFn: (input: { id: string; filter?: unknown }) =>
      api<DiscoverResult>(`/api/endpoints/${input.id}/discover`, { method: 'POST', json: { filter: input.filter } }),
  });
}

export function useOverrides(id: string | undefined) {
  return useQuery({
    queryKey: keys.overrides(id ?? ''),
    queryFn: () =>
      api<{
        providers: { raw_key: string; display_name: string; merged_into: string; source: string }[];
        families: { model_id: string; family_key: string; canon: string; source: string }[];
      }>(`/api/endpoints/${id}/overrides`),
    enabled: !!id,
  });
}

export function useSetProviderOverride() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { endpointId: string; raw_key: string; display_name?: string; merged_into?: string }) =>
      api(`/api/endpoints/${input.endpointId}/overrides/provider`, {
        method: 'PUT',
        json: { raw_key: input.raw_key, display_name: input.display_name ?? '', merged_into: input.merged_into ?? '' },
      }),
    onSuccess: (_d, v) => {
      void qc.invalidateQueries({ queryKey: keys.overrides(v.endpointId) });
      void qc.invalidateQueries({ queryKey: ['run'] });
    },
  });
}

export function useSetFamilyOverride() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: { endpointId: string; model_id: string; family_key: string; canon?: string }) =>
      api(`/api/endpoints/${input.endpointId}/overrides/family`, {
        method: 'PUT',
        json: { model_id: input.model_id, family_key: input.family_key, canon: input.canon ?? '' },
      }),
    onSuccess: (_d, v) => {
      void qc.invalidateQueries({ queryKey: keys.overrides(v.endpointId) });
      void qc.invalidateQueries({ queryKey: ['run'] });
    },
  });
}

export function useRuns(kind?: string, limit = 50) {
  return useQuery({
    queryKey: [...keys.runs(kind), limit],
    queryFn: () => api<{ runs: Run[]; total: number }>(`/api/runs?limit=${limit}${kind ? `&kind=${kind}` : ''}`),
    refetchInterval: (q) => (q.state.data?.runs.some((r) => r.running) ? 3000 : false),
  });
}

export function useRun(id: string | undefined) {
  return useQuery({
    queryKey: keys.run(id ?? ''),
    queryFn: () => api<{ run: Run; children: Run[]; running: boolean; latest: RunEvent | null }>(`/api/runs/${id}`),
    enabled: !!id,
    refetchInterval: (q) => (q.state.data?.running ? 3000 : false),
  });
}

export function useStartRun() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: Record<string, unknown>) => api<{ id: string }>('/api/runs', { method: 'POST', json: input }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['runs'] }),
  });
}

export function useCancelRun() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api(`/api/runs/${id}/cancel`, { method: 'POST' }),
    onSuccess: (_d, id) => {
      void qc.invalidateQueries({ queryKey: keys.run(id) });
      void qc.invalidateQueries({ queryKey: ['runs'] });
    },
  });
}

export function useDeleteRun() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api(`/api/runs/${id}`, { method: 'DELETE' }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['runs'] }),
  });
}

export function useSummary(id: string | undefined, enabled = true) {
  return useQuery({
    queryKey: keys.summary(id ?? ''),
    queryFn: () => api<RunSummaryDoc>(`/api/runs/${id}/summary`),
    enabled: !!id && enabled,
  });
}
export function useFamilies(id: string | undefined, order: 'composite' | 'stability', enabled = true) {
  return useQuery({
    queryKey: keys.families(id ?? '', order),
    queryFn: () =>
      api<{ families: FamilyStat[]; unavailable: FamilyStat[] }>(`/api/runs/${id}/families?order=${order}`),
    enabled: !!id && enabled,
  });
}
export function useFamily(id: string | undefined, key: string | undefined) {
  return useQuery({
    queryKey: ['run', id ?? '', 'family', key ?? ''],
    queryFn: () =>
      api<{ family: FamilyStat; members: ModelStat[] }>(`/api/runs/${id}/families/${encodeURIComponent(key ?? '')}`),
    enabled: !!id && !!key,
  });
}
export function useProviders(id: string | undefined, enabled = true) {
  return useQuery({
    queryKey: keys.providers(id ?? ''),
    queryFn: () =>
      api<{
        providers: ProviderStat[];
        field: string;
        rationale: string;
        suggestions?: { keys: string[]; evidence: string }[];
      }>(`/api/runs/${id}/providers`),
    enabled: !!id && enabled,
  });
}
export function useModels(id: string | undefined, query: string, enabled = true) {
  return useQuery({
    queryKey: keys.models(id ?? '', query),
    queryFn: () =>
      api<{ models: ModelStat[]; total: number; page: number; size: number }>(`/api/runs/${id}/models?${query}`),
    enabled: !!id && enabled,
    placeholderData: (prev) => prev,
  });
}
export function useModel(id: string | undefined, model: string | undefined) {
  return useQuery({
    queryKey: keys.model(id ?? '', model ?? ''),
    queryFn: () =>
      api<{ model: ModelStat; history: Attempt[] | null }>(`/api/runs/${id}/models/${encodePath(model ?? '')}`),
    enabled: !!id && !!model,
  });
}
export function useFailures(id: string | undefined, enabled = true) {
  return useQuery({
    queryKey: keys.failures(id ?? ''),
    queryFn: () => api<{ groups: FailureGroup[] }>(`/api/runs/${id}/failures`),
    enabled: !!id && enabled,
  });
}
export function useLatency(id: string | undefined, enabled = true) {
  return useQuery({
    queryKey: keys.latency(id ?? ''),
    queryFn: () =>
      api<{ run: Run; stats: LatencyModelStats[] | null; samples: LatencySample[] }>(`/api/runs/${id}/latency`),
    enabled: !!id && enabled,
    retry: false,
  });
}
export function useLoad(id: string | undefined, enabled = true) {
  return useQuery({
    queryKey: keys.load(id ?? ''),
    queryFn: () =>
      api<{ run: Run; summary: { load: LoadLevel[]; model: string; mode: string }; requests: unknown[] }>(
        `/api/runs/${id}/load`,
      ),
    enabled: !!id && enabled,
    retry: false,
  });
}
export function useCapabilities(id: string | undefined, enabled = true) {
  return useQuery({
    queryKey: keys.capabilities(id ?? ''),
    queryFn: () => api<{ run: Run; matrix: CapMatrix }>(`/api/runs/${id}/capabilities`),
    enabled: !!id && enabled,
    retry: false,
  });
}
export function useQuality(id: string | undefined, enabled = true) {
  return useQuery({
    queryKey: keys.quality(id ?? ''),
    queryFn: () => api<{ run: Run; scores: QualityScore[] | null; results: unknown[] }>(`/api/runs/${id}/quality`),
    enabled: !!id && enabled,
    retry: false,
  });
}
export function useCompare(a: string | undefined, b: string | undefined) {
  return useQuery({
    queryKey: keys.compare(a ?? '', b ?? ''),
    queryFn: () => api<CompareDoc>(`/api/compare?a=${a}&b=${b}`),
    enabled: !!a && !!b,
  });
}
