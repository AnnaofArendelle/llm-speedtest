import { enUS, zhCN } from '@mui/x-data-grid/locales';
import type { SxProps, Theme } from '@mui/material/styles';

export const gridDefaults = {
  density: 'compact',
  disableRowSelectionOnClick: true,
  pageSizeOptions: [25, 50, 100],
  initialState: { pagination: { paginationModel: { pageSize: 25 } } },
} as const;

export function gridLocale(lang: string) {
  return (lang.startsWith('zh') ? zhCN : enUS).components.MuiDataGrid.defaultProps.localeText;
}

export const gridSx: SxProps<Theme> = {
  border: 0,
  '& .MuiDataGrid-columnHeaderTitle': { fontWeight: 700 },
  '& .MuiDataGrid-cell': { display: 'flex', alignItems: 'center' },
  '& .MuiDataGrid-row:hover': { cursor: 'pointer' },
};
