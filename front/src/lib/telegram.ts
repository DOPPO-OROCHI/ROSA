import { request } from "./api";

type TelegramWebApp = {
  initData?: string;
  ready?: () => void;
  expand?: () => void;
};

declare global {
  interface Window {
    Telegram?: {
      WebApp?: TelegramWebApp;
    };
  }
}

export function getTelegramWebApp(): TelegramWebApp | null {
  if (typeof window === "undefined") {
    return null;
  }
  return window.Telegram?.WebApp ?? null;
}

export function getTelegramInitData(): string {
  return getTelegramWebApp()?.initData?.trim() ?? "";
}

export function isTelegramWebApp(): boolean {
  return getTelegramInitData().length > 0;
}

export async function authorizeWithTelegram(): Promise<boolean> {
  const webApp = getTelegramWebApp();
  const initData = webApp?.initData?.trim() ?? "";
  if (!initData) {
    return false;
  }

  webApp?.ready?.();
  webApp?.expand?.();

  await request<{ user_id: number; tg_id: number }>("/auth/telegram", {
    method: "POST",
    body: JSON.stringify({ initData }),
  });

  return true;
}
