import { createBrowserRouter } from 'react-router';
import { Layout } from './Layout';

const page = (load: () => Promise<{ default: React.ComponentType }>) => async () => {
  const m = await load();
  return { Component: m.default };
};

export const router = createBrowserRouter([
  {
    path: '/',
    element: <Layout />,
    children: [
      { index: true, lazy: page(() => import('@/pages/Dashboard')) },
      { path: 'new', lazy: page(() => import('@/pages/NewRun')) },
      { path: 'endpoints', lazy: page(() => import('@/pages/Endpoints')) },
      { path: 'runs', lazy: page(() => import('@/pages/Runs')) },
      { path: 'runs/:id', lazy: page(() => import('@/pages/RunDetail')) },
      { path: 'compare', lazy: page(() => import('@/pages/Compare')) },
      { path: 'settings', lazy: page(() => import('@/pages/Settings')) },
    ],
  },
]);
