import { ref } from 'vue';
import { en } from './en';
import { id } from './id';

export type Locale = 'en' | 'id';

const STORAGE_KEY = 'acis_language_pref';

// Saved or browser default language
const savedLocale = (localStorage.getItem(STORAGE_KEY) as Locale) || 'en';
export const currentLocale = ref<Locale>(savedLocale === 'id' ? 'id' : 'en');

export const dictionaries = {
  en,
  id,
};

/**
 * Access nested dictionary keys like 'login.title' or 'dashboard.wallets.title'
 */
export function t(path: string, params?: Record<string, string | number>): string {
  const keys = path.split('.');
  const dict = dictionaries[currentLocale.value] || dictionaries.en;
  
  let current: any = dict;
  for (const key of keys) {
    if (current && typeof current === 'object' && key in current) {
      current = current[key];
    } else {
      // Fallback to English dictionary if key missing in current locale
      let fallbackCurrent: any = dictionaries.en;
      for (const fbKey of keys) {
        if (fallbackCurrent && typeof fallbackCurrent === 'object' && fbKey in fallbackCurrent) {
          fallbackCurrent = fallbackCurrent[fbKey];
        } else {
          return path;
        }
      }
      current = fallbackCurrent;
      break;
    }
  }

  if (typeof current !== 'string') {
    return path;
  }

  let text = current;
  if (params) {
    for (const [pKey, pVal] of Object.entries(params)) {
      text = text.replace(new RegExp(`{${pKey}}`, 'g'), String(pVal));
    }
  }

  return text;
}

export function setLocale(newLocale: Locale): void {
  currentLocale.value = newLocale;
  localStorage.setItem(STORAGE_KEY, newLocale);
}

export function useI18n() {
  return {
    locale: currentLocale,
    t,
    setLocale,
  };
}
