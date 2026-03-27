export type TelegramWebApp = {
  ready: () => void;
  expand: () => void;
  requestFullscreen?: () => void | Promise<void>;
  exitFullscreen?: () => void | Promise<void>;
  isFullscreen?: boolean;
  isExpanded?: boolean;
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
  const webApp = getTelegramWebApp();
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

export function getTelegramWebApp(): TelegramWebApp | null {
  return window.Telegram?.WebApp ?? null;
}

export async function requestMiniAppFullscreen(): Promise<boolean> {
  const webApp = getTelegramWebApp();
  try {
    webApp?.expand();
    if (typeof webApp?.requestFullscreen === "function") {
      await webApp.requestFullscreen();
      return true;
    }
  } catch {
    // fall through to browser fullscreen
  }

  const root = document.documentElement;
  if (typeof root.requestFullscreen === "function") {
    await root.requestFullscreen();
    return true;
  }
  return false;
}

export async function exitMiniAppFullscreen(): Promise<boolean> {
  const webApp = getTelegramWebApp();
  try {
    if (typeof webApp?.exitFullscreen === "function") {
      await webApp.exitFullscreen();
      return true;
    }
  } catch {
    // fall through to browser fullscreen exit
  }

  if (typeof document.exitFullscreen === "function" && document.fullscreenElement) {
    await document.exitFullscreen();
    return true;
  }
  return false;
}

export function isMiniAppFullscreen(): boolean {
  const webApp = getTelegramWebApp();
  return Boolean(document.fullscreenElement || webApp?.isFullscreen);
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
