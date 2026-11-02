import { enUS, zhCN } from '@mui/x-data-grid/locales';
import { alpha, type SxProps, type Theme } from '@mui/material/styles';

export const gridDefaults = {
  density: 'compact',
  disableRowSelectionOnClick: true,
  pageSizeOptions: [25, 50, 100],
  initialState: { pagination: { paginationModel: { pageSize: 25 } } },
  // happy-dom has no layout, so the grid would render zero rows under test.
  ...(import.meta.env.MODE === 'test' ? { disableVirtualization: true } : {}),
} as const;

export function gridLocale(lang: string) {
  return (lang.startsWith('zh') ? zhCN : enUS).components.MuiDataGrid.defaultProps.localeText;
}

/** Shared DataGrid look: tonal header, hairline rows, tinted hover. */
export const gridSx: SxProps<Theme> = (theme) => ({
  border: 0,
  '--DataGrid-containerBackground': theme.palette.mode === 'dark' ? '#182033' : '#F3F5F9',
  '--DataGrid-rowBorderColor': theme.palette.divider,
  '& .MuiDataGrid-columnHeaderTitle': { fontWeight: 700, fontSize: 12.5, letterSpacing: 0.2 },
  '& .MuiDataGrid-cell': { display: 'flex', alignItems: 'center' },
  '& .MuiDataGrid-row:hover': {
    cursor: 'pointer',
    backgroundColor: alpha(theme.palette.primary.main, theme.palette.mode === 'dark' ? 0.08 : 0.05),
  },
  '& .MuiDataGrid-footerContainer': { borderTop: `1px solid ${theme.palette.divider}` },
});

/** gridSx plus a non-clickable row cursor. */
export const gridSxStatic: SxProps<Theme> = [
  gridSx,
  { '& .MuiDataGrid-row:hover': { cursor: 'default' } },
] as SxProps<Theme>;
