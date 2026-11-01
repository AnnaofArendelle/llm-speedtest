import type { Run } from '@/api/types';

export interface TabProps {
  runId: string;
  run: Run;
  onModel: (id: string) => void;
  onFamily: (key: string) => void;
  goTab: (tab: string, params?: Record<string, string>) => void;
}
