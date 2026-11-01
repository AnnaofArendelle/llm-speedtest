import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { useReasons } from '@/api/hooks';
import type { ReasonInfo } from '@/api/types';

export interface ReasonLookup {
  info: (code: string) => ReasonInfo | undefined;
  name: (code: string) => string;
  desc: (code: string) => string;
  fix: (code: string) => string;
  all: ReasonInfo[];
}

export function useReasonLookup(): ReasonLookup {
  const { data } = useReasons();
  const { i18n } = useTranslation();
  const en = i18n.language.startsWith('en');
  return useMemo(() => {
    const list = data ?? [];
    const map = new Map(list.map((r) => [r.code, r]));
    return {
      all: list,
      info: (code) => map.get(code),
      name: (code) => {
        const r = map.get(code);
        if (!r) return code === 'ok' ? 'OK' : code;
        return en ? r.name_en || r.name : r.name;
      },
      desc: (code) => {
        const r = map.get(code);
        return r ? (en ? r.description_en : r.description) : '';
      },
      fix: (code) => {
        const r = map.get(code);
        return r ? (en ? r.fix_en : r.fix) : '';
      },
    };
  }, [data, en]);
}
