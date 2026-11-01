import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import { zh } from './locales/zh';
import { en } from './locales/en';
import { storage } from '@/lib/storage';

const saved = storage.get('lang');
const browser = typeof navigator !== 'undefined' && navigator.language.toLowerCase().startsWith('zh') ? 'zh' : 'en';

void i18n.use(initReactI18next).init({
  resources: { zh: { translation: zh }, en: { translation: en } },
  lng: saved === 'zh' || saved === 'en' ? saved : browser,
  fallbackLng: 'en',
  interpolation: { escapeValue: false },
});

export default i18n;
