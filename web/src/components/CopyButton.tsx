import { useState } from 'react';
import { IconButton, Tooltip } from '@mui/material';
import ContentCopyRounded from '@mui/icons-material/ContentCopyRounded';
import CheckRounded from '@mui/icons-material/CheckRounded';
import { useTranslation } from 'react-i18next';

export function CopyButton({ text, size = 'small' }: { text: string; size?: 'small' | 'medium' }) {
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      /* clipboard unavailable */
    }
  };
  return (
    <Tooltip title={copied ? t('common.copied') : t('common.copy')}>
      <IconButton size={size} onClick={() => void copy()} aria-label={t('common.copy')}>
        {copied ? <CheckRounded fontSize="inherit" color="success" /> : <ContentCopyRounded fontSize="inherit" />}
      </IconButton>
    </Tooltip>
  );
}
