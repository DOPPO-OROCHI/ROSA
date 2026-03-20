export type TelegramWebApp = {
  ready: () => void;
  expand: () => void;
  disableVerticalSwipes?: () => void;
  setHeaderColor?: (color: string) => void;
  setBackgroundColor?: (color: string) => void;
  initData?: string;
};

declare global {
  interface Window {
    Telegram?: {
      WebApp?: TelegramWebApp;
    };
  }
}

export function bootstrapTelegramWebApp(): TelegramWebApp | null {
  const webApp = window.Telegram?.WebApp ?? null;
  if (!webApp) {
    return null;
  }

  webApp.ready();
  webApp.expand();
  webApp.disableVerticalSwipes?.();
  webApp.setHeaderColor?.("#11151b");
  webApp.setBackgroundColor?.("#0c0d10");
  return webApp;
}

function extractTgDataFromUrl(): string {
  const fromQuery = new URLSearchParams(window.location.search).get("tgWebAppData") ?? "";
  if (fromQuery) {
    return decodeURIComponent(fromQuery);
  }
  const hash = window.location.hash.startsWith("#") ? window.location.hash.slice(1) : window.location.hash;
  const fromHash = new URLSearchParams(hash).get("tgWebAppData") ?? "";
  if (fromHash) {
    return decodeURIComponent(fromHash);
  }
  return "";
}

export function getTelegramInitData(): string {
  const fromWebApp = window.Telegram?.WebApp?.initData ?? "";
  if (fromWebApp) {
    sessionStorage.setItem("tg_init_data", fromWebApp);
    return fromWebApp;
  }

  const fromUrl = extractTgDataFromUrl();
  if (fromUrl) {
    sessionStorage.setItem("tg_init_data", fromUrl);
    return fromUrl;
  }

  return sessionStorage.getItem("tg_init_data") ?? "";
}
