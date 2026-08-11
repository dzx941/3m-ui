import React, { createContext, useContext, useState } from 'react';
import { messages, type Locale } from './messages';

interface I18nContextProps {
  locale: Locale;
  setLocale: (locale: Locale) => void;
  t: (key: string) => string;
}

const I18nContext = createContext<I18nContextProps>({
  locale: 'zh-CN',
  setLocale: () => {},
  t: (key: string) => messages['zh-CN'][key] || key,
});

export const I18nProvider: React.FC<React.PropsWithChildren> = ({ children }) => {
  const [locale, setLocaleState] = useState<Locale>(() => {
    const saved = localStorage.getItem('3m-ui.locale') as Locale | null;
    return saved === 'en-US' || saved === 'zh-CN' ? saved : 'zh-CN';
  });

  const setLocale = (l: Locale) => {
    setLocaleState(l);
    localStorage.setItem('3m-ui.locale', l);
  };

  const t = (key: string) => messages[locale][key] || messages['zh-CN'][key] || key;

  return <I18nContext.Provider value={{ locale, setLocale, t }}>{children}</I18nContext.Provider>;
};

// eslint-disable-next-line react/only-export-components
export const useI18n = () => useContext(I18nContext);
