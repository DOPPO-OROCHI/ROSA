import { useEffect, useState } from "react";

type Props = {
  message: string;
  nonce?: number;
};

function normalizeBattleMessage(message: string): string {
  if (!message) {
    return "";
  }

  const trimmed = message.trim();
  try {
    const parsed = JSON.parse(trimmed) as { error?: string; message?: string };
    return parsed.error || parsed.message || trimmed;
  } catch {
    return trimmed;
  }
}

export function BattleInfoToast({ message, nonce = 0 }: Props) {
  const [visible, setVisible] = useState(false);
  const [renderedMessage, setRenderedMessage] = useState("");

  useEffect(() => {
    if (!message) {
      setVisible(false);
      return;
    }

    setRenderedMessage(normalizeBattleMessage(message));
    setVisible(true);

    const hideId = window.setTimeout(() => {
      setVisible(false);
    }, 1900);

    return () => window.clearTimeout(hideId);
  }, [message, nonce]);

  if (!renderedMessage) {
    return null;
  }

  return (
    <div className={`battle-info-toast ${visible ? "battle-info-toast--visible" : "battle-info-toast--hidden"}`} aria-live="polite">
      {renderedMessage}
    </div>
  );
}
