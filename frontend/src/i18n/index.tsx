import React, { createContext, useContext, useState } from 'react';
import { messages, type Locale } from './messages';

interface I18nContextProps {
  locale: Locale;
  setLocale: (locale: Locale) => void;
  t: (key: string) => string;
}

const fallbackMessages: Record<Locale, Record<string, string>> = {
  'zh-CN': {
    'nodes.protocol': '协议',
    'nodes.listen': '监听地址',
    'nodes.status': '状态',
    'nodes.actions': '操作',
    'nodes.rule': '入站规则',
    'nodes.rulePlaceholder': '可选规则',
    'nodes.proxy': '代理',
    'nodes.proxyPlaceholder': '可选上游代理',
    'nodes.protocolHint': '节点由 Mihomo Listener 配置驱动，保存后会自动生成客户端代理配置。',
    'nodes.clientConfigHint': '以下链接由当前节点自动生成，可直接提供给对应客户端。',
    'nodes.mihomoClash': 'Mihomo / Clash',
    'nodes.singbox': 'sing-box',
    'nodes.shadowrocket': 'Shadowrocket',
  },
  'en-US': {
    'nodes.protocol': 'Protocol',
    'nodes.listen': 'Listen Address',
    'nodes.status': 'Status',
    'nodes.actions': 'Actions',
    'nodes.rule': 'Inbound Rule',
    'nodes.rulePlaceholder': 'Optional rule',
    'nodes.proxy': 'Proxy',
    'nodes.proxyPlaceholder': 'Optional upstream proxy',
    'nodes.protocolHint': 'Nodes are driven by Mihomo Listener configuration and automatically generate client proxy configurations after saving.',
    'nodes.clientConfigHint': 'These links are generated from the current node and can be provided directly to the corresponding client.',
    'nodes.mihomoClash': 'Mihomo / Clash',
    'nodes.singbox': 'sing-box',
    'nodes.shadowrocket': 'Shadowrocket',
  },
};

const I18nContext = createContext<I18nContextProps>({
  locale: 'zh-CN',
  setLocale: () => {},
  t: (key: string) => messages['zh-CN'][key] || fallbackMessages['zh-CN'][key] || key,
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

  const t = (key: string) => messages[locale][key] || fallbackMessages[locale][key] || messages['zh-CN'][key] || fallbackMessages['zh-CN'][key] || key;

  return <I18nContext.Provider value={{ locale, setLocale, t }}>{children}</I18nContext.Provider>;
};

// eslint-disable-next-line react/only-export-components
export const useI18n = () => useContext(I18nContext);
