export type TelegramWebApp = {
  ready: () => void;
  expand: () => void;
  disableVerticalSwipes?: () => void;
  setHeaderColor?: (color: string) => void;
  setBackgroundColor?: (color: string) => void;
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
