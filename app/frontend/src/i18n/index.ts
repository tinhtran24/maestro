import { en, type MessageKey } from "./en";

type Dict = Record<string, string>;

const dictionaries: Record<string, Dict> = {
  en,
};

export function createTranslator(locale = "en") {
  const active = dictionaries[locale] ?? en;
  return (key: MessageKey, vars?: Record<string, string | number>) => {
    let value = active[key] ?? en[key] ?? key;
    if (!vars) return value;
    for (const [name, replacement] of Object.entries(vars)) {
      value = value.replace(new RegExp(`\\{${name}\\}`, "g"), String(replacement));
    }
    return value;
  };
}

export function useT(locale = "en") {
  return createTranslator(locale);
}

export type { MessageKey };
