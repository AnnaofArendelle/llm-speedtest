import { useState } from 'react';
import {
  Box,
  Button,
  FormControl,
  InputLabel,
  LinearProgress,
  MenuItem,
  Select,
  Stack,
  Typography,
} from '@mui/material';
import DownloadRounded from '@mui/icons-material/DownloadRounded';
import DescriptionOutlined from '@mui/icons-material/DescriptionOutlined';
import DataObjectRounded from '@mui/icons-material/DataObjectRounded';
import TableChartOutlined from '@mui/icons-material/TableChartOutlined';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { apiText, withToken } from '@/api/client';
import { Section } from '@/components/Section';
import { CopyButton } from '@/components/CopyButton';
import type { TabProps } from './types';

export default function ExportTab({ runId, run }: TabProps) {
  const { t, i18n } = useTranslation();
  const [lang, setLang] = useState(i18n.language.startsWith('zh') ? 'zh' : 'en');
  const [preview, setPreview] = useState(false);
  const md = useQuery({
    queryKey: ['run', runId, 'report', lang],
    queryFn: () => apiText(`/api/runs/${runId}/report.md?lang=${lang}`),
    enabled: preview,
  });
  const hasScan = run.kind === 'scan' || run.kind === 'suite';
  return (
    <Stack spacing={2}>
      <Typography variant="body2" color="text.secondary">
        {t('run.export.intro')}
      </Typography>
      <Section title={t('common.export')}>
        <Stack
          direction={{ xs: 'column', sm: 'row' }}
          spacing={1.5}
          alignItems={{ sm: 'center' }}
          flexWrap="wrap"
          useFlexGap
        >
          <FormControl size="small" sx={{ minWidth: 160 }}>
            <InputLabel>{t('run.export.lang')}</InputLabel>
            <Select label={t('run.export.lang')} value={lang} onChange={(e) => setLang(e.target.value)}>
              <MenuItem value="zh">中文</MenuItem>
              <MenuItem value="en">English</MenuItem>
            </Select>
          </FormControl>
          <Button
            variant="contained"
            disableElevation
            startIcon={<DescriptionOutlined />}
            component="a"
            href={withToken(`/api/runs/${runId}/report.md?lang=${lang}&download=1`)}
            download
          >
            {t('run.export.markdown')}
          </Button>
          {hasScan && (
            <>
              <Button
                variant="outlined"
                startIcon={<DataObjectRounded />}
                component="a"
                href={withToken(`/api/runs/${runId}/export.json`)}
                download
              >
                {t('run.export.json')}
              </Button>
              <Button
                variant="outlined"
                startIcon={<TableChartOutlined />}
                component="a"
                href={withToken(`/api/runs/${runId}/export.csv`)}
                download
              >
                {t('run.export.csv')}
              </Button>
            </>
          )}
          <Button variant="text" startIcon={<DownloadRounded />} onClick={() => setPreview(true)}>
            {t('run.export.preview')}
          </Button>
        </Stack>
      </Section>
      {preview && (
        <Section title={t('run.export.markdown')} action={md.data ? <CopyButton text={md.data} /> : undefined} noPad>
          {md.isLoading && <LinearProgress />}
          {md.data && (
            <Box
              component="pre"
              className="mono"
              sx={{
                m: 0,
                p: 2,
                fontSize: 12.5,
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-word',
                maxHeight: '70vh',
                overflow: 'auto',
              }}
            >
              {md.data}
            </Box>
          )}
        </Section>
      )}
    </Stack>
  );
}
