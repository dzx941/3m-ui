import React, { createContext, useContext } from 'react';
import { messages, type Locale } from './messages';

const I18nContext = createContext({ locale: 'zh-CN' as Locale, t: (key: string) => messages['zh-CN'][key] || key });

export const I18nProvider: React.FC<React.PropsWithChildren<{ locale?: Locale }>> = ({ locale = 'zh-CN', children }) => {
  const t = (key: string) => messages[locale][key] || messages['zh-CN'][key] || key;
  return <I18nContext.Provider value={{ locale, t }}>{children}</I18nContext.Provider>;
};

// eslint-disable-next-line react/only-export-components
export const useI18n = () => useContext(I18nContext);
