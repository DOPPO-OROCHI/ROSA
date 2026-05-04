import { useEffect, useRef, useState } from "react";
import { GameModePanel } from "./components/GameModePanel";
import { InventoryScreen } from "./components/InventoryScreen";
import { MainMenu } from "./components/MainMenu";
import { MatchFoundPanel } from "./components/MatchFoundPanel";
import { BattleScreen } from "./components/battle/BattleScreen";
import { BATTLE_MUSIC_TRACKS, BattleMusicManager } from "./lib/battle_music";
import { request, resolveCardAssetVariantSrc, resolveHeroAssetVariantSrc, setDevSessionToken } from "./lib/api";
import { uniqueAssetUrls, warmAssetUrlsInBackground } from "./lib/asset_preload";
import type {
  BattleCard,
  BuffCard,
  CardsResponse,
  DeckEntry,
  DeckResponse,
  Hero,
  HeroesResponse,
  JoinQueueResponse,
  MeResponse,
  QueueState,
  QueueStatusResponse,
} from "./types";
import type { MaskedBattleMatchState } from "./components/battle/types";

const QUICK_USERS = ["dev", "roman", "test", "rosa"];
const MENU_MUSIC_VOLUME = 0.35;
const MENU_MUSIC_MATCH_FOUND_VOLUME = MENU_MUSIC_VOLUME * 0.5;
const BATTLE_MUSIC_VOLUME = 0.28;
const MUSIC_FADE_MS = 420;
const COMMON_MENU_WARM_ASSET_URLS = [
  "/assets/ui/pictures/backgrounds/menu/image.png",
  "/assets/ui/pictures/backgrounds/inventory/image.png",
  "/assets/ui/pictures/backgrounds/battle/image.png",
  "/assets/ui/pictures/backgrounds/queue_search/queue_search_01.png",
  "/assets/ui/pictures/boards/battle/image.png",
  "/assets/ui/pictures/panels/game_mode/image.png",
  "/assets/ui/pictures/panels/game_mode_background/image.png",
  "/assets/ui/pictures/panels/shop/image.png",
  "/assets/ui/sounds/ui/click.mp3",
  "/assets/ui/sounds/ui/match_found.mp3",
  "/assets/ui/sounds/combat/impact.mp3",
  "/assets/ui/sounds/music/menu_music.mp3",
  "/assets/ui/pictures/icons/status/atk_cd_down.png",
  "/assets/ui/pictures/icons/status/atk_cd_up.png",
  "/assets/ui/pictures/icons/status/buff_atk.png",
  "/assets/ui/pictures/icons/status/buff_atk_hp.png",
  "/assets/ui/pictures/icons/status/buff_hp.png",
  "/assets/ui/pictures/icons/status/damage.png",
  "/assets/ui/pictures/icons/status/death.png",
  "/assets/ui/pictures/icons/status/disarm.png",
  "/assets/ui/pictures/icons/status/dot.png",
  "/assets/ui/pictures/icons/status/heal_over_time.png",
  "/assets/ui/pictures/icons/status/no_heal.png",
  "/assets/ui/pictures/icons/status/reflect_shield.png",
  "/assets/ui/pictures/icons/status/shield.png",
  "/assets/ui/pictures/icons/status/silence.png",
  "/assets/ui/pictures/icons/status/skill.png",
  "/assets/ui/pictures/icons/status/skill_cd_down.png",
  "/assets/ui/pictures/icons/status/skill_cd_up.png",
  "/assets/ui/pictures/icons/status/stun.png",
  "/assets/ui/pictures/icons/status/summon.png",
  "/assets/ui/pictures/icons/status/vulnerable.png",
];

type Screen = "menu" | "inventory" | "battle";

type DevAuthResponse = {
  user_id: number;
  username?: string;
  token?: string;
};

type TelegramWebAppWindow = Window & {
  Telegram?: {
    WebApp?: {
      initData?: string;
      ready?: () => void;
      expand?: () => void;
      isFullscreen?: boolean;
      requestFullscreen?: () => void;
      exitFullscreen?: () => void;
      onEvent?: (eventType: string, eventHandler: () => void) => void;
      offEvent?: (eventType: string, eventHandler: () => void) => void;
    };
  };
};

function addWarmCardAssets(urls: string[], kind: "battle" | "buff", templateId: string, includeBattleSfx: boolean) {
  urls.push(resolveCardAssetVariantSrc(kind, templateId, "view"));
  urls.push(resolveCardAssetVariantSrc(kind, templateId, "full_art"));

  if (kind !== "battle") {
    return;
  }

  urls.push(resolveCardAssetVariantSrc(kind, templateId, "on_table"));

  if (!includeBattleSfx) {
    return;
  }

  urls.push(`/assets/cards/battle/${templateId}/sfx/summon/sound.mp3`);
  urls.push(`/assets/cards/battle/${templateId}/sfx/death/sound.mp3`);
  urls.push(`/assets/cards/battle/${templateId}/sfx/spell/sound.mp3`);
}

function collectMenuWarmAssetUrls(heroes: Hero[], battleCards: BattleCard[], buffCards: BuffCard[], deckEntries: DeckEntry[]) {
  const urls = [...COMMON_MENU_WARM_ASSET_URLS, ...BATTLE_MUSIC_TRACKS.slice(0, 2)];
  const deckKeys = new Set(deckEntries.filter((entry) => entry.count > 0).map((entry) => `${entry.kind}:${entry.template_id}`));

  heroes.forEach((hero) => {
    urls.push(resolveHeroAssetVariantSrc(hero.hero_code, "view"));
    urls.push(resolveHeroAssetVariantSrc(hero.hero_code, "full_art"));
    urls.push(resolveHeroAssetVariantSrc(hero.hero_code, "battle_icon"));
  });

  battleCards.forEach((card) => {
    addWarmCardAssets(urls, "battle", card.template_id, deckKeys.has(`battle:${card.template_id}`));
  });

  buffCards.forEach((card) => {
    addWarmCardAssets(urls, "buff", card.template_id, deckKeys.has(`buff:${card.template_id}`));
  });

  return uniqueAssetUrls(urls);
}

export function App() {
  const clickAudioRef = useRef<HTMLAudioElement | null>(null);
  const matchFoundAudioRef = useRef<HTMLAudioElement | null>(null);
  const menuMusicRef = useRef<HTMLAudioElement | null>(null);
  const battleMusicManagerRef = useRef<BattleMusicManager | null>(null);
  const menuMusicFadeFrameRef = useRef<number | null>(null);
  const musicSyncGenerationRef = useRef(0);
  const screenRef = useRef<Screen>("menu");
  const musicEnabledRef = useRef(true);
  const queueStateRef = useRef<QueueState>("idle");
  const prevQueueStateRef = useRef<QueueState>("idle");
  const [screen, setScreen] = useState<Screen>("menu");
  const [username, setUsername] = useState("dev");
  const [userId, setUserId] = useState("");
  const [me, setMe] = useState<MeResponse | null>(null);
  const [heroes, setHeroes] = useState<Hero[]>([]);
  const [battleCards, setBattleCards] = useState<BattleCard[]>([]);
  const [buffCards, setBuffCards] = useState<BuffCard[]>([]);
  const [deckEntries, setDeckEntries] = useState<DeckEntry[]>([]);
  const [draftDeckEntries, setDraftDeckEntries] = useState<DeckEntry[]>([]);
  const [busy, setBusy] = useState(true);
  const [error, setError] = useState("");
  const [selectedHeroCode, setSelectedHeroCode] = useState("");
  const [heroPickerOpen, setHeroPickerOpen] = useState(false);
  const [gameModeOpen, setGameModeOpen] = useState(false);
  const [queueBusy, setQueueBusy] = useState(false);
  const [queueError, setQueueError] = useState("");
  const [queueStatus, setQueueStatus] = useState<QueueStatusResponse>({ state: "idle" });
  const [activeMatchId, setActiveMatchId] = useState<number | null>(null);
  const [matchFoundVerdict, setMatchFoundVerdict] = useState<"idle" | "accepted" | "declined_self" | "declined_opponent" | "countdown">("idle");
  const [matchFoundCountdown, setMatchFoundCountdown] = useState(3);
  const [acceptedByMeLatch, setAcceptedByMeLatch] = useState(false);
  const [musicEnabled, setMusicEnabled] = useState(true);
  const [fullscreenActive, setFullscreenActive] = useState(false);

  const selectedHero = heroes.find((hero) => hero.hero_code === selectedHeroCode) ?? null;
  const deckCardCount = draftDeckEntries.reduce((total, entry) => total + entry.count, 0);
  const canJoinQueue = Boolean(selectedHeroCode) && deckCardCount === 20;
  const queueHint = !selectedHeroCode && deckCardCount !== 20
    ? "НЕТ ГЕРОЯ | НЕПОЛНАЯ ДЕКА"
    : !selectedHeroCode
      ? "НЕТ ГЕРОЯ"
      : deckCardCount !== 20
        ? "НЕПОЛНАЯ ДЕКА"
        : "";
  const queueSearching = queueStatus.state === "searching" || queueStatus.state === "pending_match";
  const queuePenalty = queueStatus.state === "penalty";
  const matchTransitionLocked = Boolean(activeMatchId) && (acceptedByMeLatch || matchFoundVerdict === "countdown");
  const startMatchLabel = !selectedHeroCode && deckCardCount !== 20
    ? "НЕТ ГЕРОЯ / НЕПОЛНАЯ ДЕКА"
    : !selectedHeroCode
      ? "НЕ ВЫБРАН ПЕРСОНАЖ"
      : deckCardCount !== 20
        ? "НЕПОЛНАЯ ДЕКА"
        : "START MATCH";
  function cancelAudioFade(frameRef: React.MutableRefObject<number | null>) {
    if (frameRef.current !== null) {
      window.cancelAnimationFrame(frameRef.current);
      frameRef.current = null;
    }
  }

  function fadeAudioTo(
    audio: HTMLAudioElement | null,
    frameRef: React.MutableRefObject<number | null>,
    targetVolume: number,
    durationMs = MUSIC_FADE_MS,
    onDone?: () => void,
  ) {
    if (!audio) {
      return;
    }

    cancelAudioFade(frameRef);

    const startVolume = audio.volume;
    if (Math.abs(startVolume - targetVolume) < 0.005) {
      audio.volume = targetVolume;
      onDone?.();
      return;
    }

    const startedAt = performance.now();
    const tick = (now: number) => {
      const progress = Math.min((now - startedAt) / durationMs, 1);
      const eased = 1 - Math.pow(1 - progress, 3);
      audio.volume = startVolume + (targetVolume - startVolume) * eased;

      if (progress < 1) {
        frameRef.current = window.requestAnimationFrame(tick);
        return;
      }

      audio.volume = targetVolume;
      frameRef.current = null;
      onDone?.();
    };

    frameRef.current = window.requestAnimationFrame(tick);
  }

  function getMenuMusicTargetVolume(nextQueueState = queueStateRef.current) {
    return nextQueueState === "pending_match"
      ? MENU_MUSIC_MATCH_FOUND_VOLUME
      : MENU_MUSIC_VOLUME;
  }

  function fadeMenuMusicTo(targetVolume: number, durationMs = MUSIC_FADE_MS, onDone?: () => void) {
    fadeAudioTo(menuMusicRef.current, menuMusicFadeFrameRef, targetVolume, durationMs, onDone);
  }

  function syncMusicPlayback(nextScreen: Screen, nextMusicEnabled: boolean, nextQueueState = queueStateRef.current) {
    const generation = musicSyncGenerationRef.current + 1;
    musicSyncGenerationRef.current = generation;
    const menuAudio = menuMusicRef.current;
    const battleMusic = battleMusicManagerRef.current;
    if (!menuAudio || !battleMusic) {
      return;
    }

    if (!nextMusicEnabled) {
      cancelAudioFade(menuMusicFadeFrameRef);
      menuAudio.pause();
      menuAudio.volume = getMenuMusicTargetVolume(nextQueueState);
      battleMusic.stop(0);
      return;
    }

    if (nextScreen === "menu" || nextScreen === "inventory") {
      if (menuAudio.paused) {
        void menuAudio.play().catch(() => undefined);
      }
      battleMusic.stop(MUSIC_FADE_MS);
      fadeMenuMusicTo(getMenuMusicTargetVolume(nextQueueState));
      return;
    }

    if (nextScreen === "battle") {
      fadeMenuMusicTo(0, MUSIC_FADE_MS, () => {
        if (musicSyncGenerationRef.current === generation && screenRef.current === "battle") {
          menuAudio.pause();
          menuAudio.currentTime = 0;
        }
      });
      battleMusic.start();
    }
  }

  async function loadQueueStatus() {
    const status = await request<QueueStatusResponse>("/queue/status");
    setQueueStatus(status);
    setQueueError("");
  }

  async function loadActiveMatch() {
    const matches = await request<MaskedBattleMatchState[]>("/matches");
    const active = matches.find((match) => !match.finished);
    if (active) {
      setActiveMatchId(active.match_id);
      if (queueStatus.state === "pending_match" || acceptedByMeLatch || matchFoundVerdict === "countdown") {
        setMatchFoundVerdict("countdown");
        return active.match_id;
      }
      setScreen("battle");
      return active.match_id;
    }
    setActiveMatchId(null);
    return null;
  }

  async function loadSession() {
    const [nextMe, nextHeroes] = await Promise.all([
      request<MeResponse>("/me"),
      request<HeroesResponse>("/heroes"),
    ]);
    setMe(nextMe);
    setHeroes(nextHeroes.heroes);
    setSelectedHeroCode(nextMe.selected_hero_code || "");
    setError("");
  }

  async function loadInventory() {
    const [cards, deck] = await Promise.all([
      request<CardsResponse>("/cards"),
      request<DeckResponse>("/deck"),
    ]);
    setBattleCards(cards.battle);
    setBuffCards(cards.buff);
    setDeckEntries(deck.entries);
    setDraftDeckEntries(deck.entries);
  }

  async function authenticateWithTelegram() {
    setDevSessionToken();

    const webApp = (window as TelegramWebAppWindow).Telegram?.WebApp;
    webApp?.ready?.();
    webApp?.expand?.();

    const initData = webApp?.initData ?? "";
    if (!initData) {
      return;
    }

    await request("/auth/telegram", {
      method: "POST",
      body: JSON.stringify({ initData }),
    });
  }

  useEffect(() => {
    authenticateWithTelegram()
      .then(loadSession)
      .then(async () => {
        await loadInventory();
        await loadQueueStatus();
        await loadActiveMatch();
      })
      .catch(() => {
        setMe(null);
        setHeroes([]);
        setBattleCards([]);
        setBuffCards([]);
        setDeckEntries([]);
        setDraftDeckEntries([]);
        setSelectedHeroCode("");
        setQueueStatus({ state: "idle" });
        setActiveMatchId(null);
        setError("Открой игру через кнопку ИГРАТЬ в Telegram-боте");
      })
      .finally(() => setBusy(false));
  }, []);

  useEffect(() => {
    screenRef.current = screen;
  }, [screen]);

  useEffect(() => {
    musicEnabledRef.current = musicEnabled;
  }, [musicEnabled]);

  useEffect(() => {
    queueStateRef.current = queueStatus.state;
  }, [queueStatus.state]);

  useEffect(() => {
    const webApp = (window as TelegramWebAppWindow).Telegram?.WebApp;
    const syncFullscreenState = () => {
      setFullscreenActive(Boolean(document.fullscreenElement || webApp?.isFullscreen));
    };

    syncFullscreenState();
    document.addEventListener("fullscreenchange", syncFullscreenState);
    webApp?.onEvent?.("fullscreenChanged", syncFullscreenState);
    return () => {
      document.removeEventListener("fullscreenchange", syncFullscreenState);
      webApp?.offEvent?.("fullscreenChanged", syncFullscreenState);
    };
  }, []);

  useEffect(() => {
    const audio = new Audio("/assets/ui/sounds/ui/click.mp3");
    audio.preload = "auto";
    clickAudioRef.current = audio;

    function handleUiClick(event: PointerEvent) {
      const target = event.target;
      if (!(target instanceof Element)) {
        return;
      }

      const interactive = target.closest(
        'button, a, input, select, textarea, summary, [role="button"], [data-ui-click]',
      );

      if (!(interactive instanceof HTMLElement)) {
        return;
      }

      if (
        interactive.matches(":disabled") ||
        interactive.getAttribute("aria-disabled") === "true"
      ) {
        return;
      }

      const baseAudio = clickAudioRef.current;
      if (!baseAudio) {
        return;
      }

      const clickAudio = baseAudio.cloneNode() as HTMLAudioElement;
      clickAudio.volume = 0.1;
      void clickAudio.play().catch(() => undefined);

      const battleMusic = battleMusicManagerRef.current;

      if (!musicEnabledRef.current) {
        return;
      }

      syncMusicPlayback(screenRef.current, musicEnabledRef.current, queueStateRef.current);

      if (battleMusic) {
        void battleMusic.unlock().then(() => {
          syncMusicPlayback(screenRef.current, musicEnabledRef.current, queueStateRef.current);
        });
      }
    }

    window.addEventListener("pointerdown", handleUiClick, { passive: true });
    return () => window.removeEventListener("pointerdown", handleUiClick);
  }, []);

  useEffect(() => {
    const audio = new Audio("/assets/ui/sounds/ui/match_found.mp3");
    audio.preload = "auto";
    audio.volume = 0.62;
    matchFoundAudioRef.current = audio;

    return () => {
      audio.pause();
      audio.currentTime = 0;
      if (matchFoundAudioRef.current === audio) {
        matchFoundAudioRef.current = null;
      }
    };
  }, []);

  useEffect(() => {
    const audio = new Audio("/assets/ui/sounds/music/menu_music.mp3");
    audio.preload = "auto";
    audio.loop = true;
    audio.volume = MENU_MUSIC_VOLUME;
    menuMusicRef.current = audio;
    syncMusicPlayback(screenRef.current, musicEnabledRef.current, queueStateRef.current);

    return () => {
      if (menuMusicFadeFrameRef.current !== null) {
        window.cancelAnimationFrame(menuMusicFadeFrameRef.current);
        menuMusicFadeFrameRef.current = null;
      }
      audio.pause();
      audio.currentTime = 0;
      if (menuMusicRef.current === audio) {
        menuMusicRef.current = null;
      }
    };
  }, []);

  useEffect(() => {
    const manager = new BattleMusicManager(BATTLE_MUSIC_TRACKS, {
      volume: BATTLE_MUSIC_VOLUME,
      crossfadeMs: 4200,
    });
    battleMusicManagerRef.current = manager;
    syncMusicPlayback(screenRef.current, musicEnabledRef.current, queueStateRef.current);

    return () => {
      manager.destroy();
      if (battleMusicManagerRef.current === manager) {
        battleMusicManagerRef.current = null;
      }
    };
  }, []);

  useEffect(() => {
    syncMusicPlayback(screen, musicEnabled, queueStatus.state);
  }, [musicEnabled, queueStatus.state, screen]);

  useEffect(() => {
    if (!me || heroes.length === 0 || battleCards.length === 0) {
      return;
    }

    const stopWarmup = warmAssetUrlsInBackground(
      collectMenuWarmAssetUrls(heroes, battleCards, buffCards, deckEntries),
      {
        concurrency: 3,
        initialDelayMs: 900,
        batchDelayMs: 80,
      },
    );

    return stopWarmup;
  }, [battleCards, buffCards, deckEntries, heroes, me]);

  useEffect(() => {
    const previousState = prevQueueStateRef.current;
    if (queueStatus.state === "pending_match" && previousState !== "pending_match") {
      const baseAudio = matchFoundAudioRef.current;
      if (baseAudio) {
        const matchFoundAudio = baseAudio.cloneNode() as HTMLAudioElement;
        matchFoundAudio.volume = baseAudio.volume;
        void matchFoundAudio.play().catch(() => undefined);
      }
    }
    prevQueueStateRef.current = queueStatus.state;
  }, [queueStatus.state]);

  useEffect(() => {
    if (!queueSearching && !queuePenalty) {
      return;
    }

    const id = window.setInterval(() => {
      void loadQueueStatus().catch(() => {
        setQueueStatus((current) => ({ ...current, state: "idle" }));
      });
      if (queueSearching) {
        void loadActiveMatch().catch(() => undefined);
      }
    }, 1000);

    return () => window.clearInterval(id);
  }, [queuePenalty, queueSearching]);

  useEffect(() => {
    if (queueStatus.state === "pending_match") {
      if (queueStatus.accepted_by_me) {
        setAcceptedByMeLatch(true);
        setMatchFoundVerdict("accepted");
      } else {
        setMatchFoundVerdict("idle");
      }
      return;
    }

    if (queueStatus.state === "searching" && acceptedByMeLatch) {
      setMatchFoundVerdict("declined_opponent");
      const id = window.setTimeout(() => {
        setMatchFoundVerdict("idle");
        setAcceptedByMeLatch(false);
      }, 3000);
      return () => window.clearTimeout(id);
    }

    if (queueStatus.state === "penalty") {
      const id = window.setTimeout(() => {
        setMatchFoundVerdict("idle");
        setAcceptedByMeLatch(false);
      }, 3000);
      return () => window.clearTimeout(id);
    }
  }, [acceptedByMeLatch, queueStatus.state, queueStatus.accepted_by_me]);

  useEffect(() => {
    if (!activeMatchId || !acceptedByMeLatch || matchFoundVerdict === "countdown") {
      return;
    }
    setMatchFoundVerdict("countdown");
  }, [acceptedByMeLatch, activeMatchId, matchFoundVerdict]);

  useEffect(() => {
    if (matchFoundVerdict !== "countdown" || !activeMatchId) {
      setMatchFoundCountdown(3);
      return;
    }

    const id = window.setInterval(() => {
      setMatchFoundCountdown((current) => {
        if (current <= 1) {
          window.clearInterval(id);
          setScreen("battle");
          setMatchFoundVerdict("idle");
          setAcceptedByMeLatch(false);
          return 0;
        }
        return current - 1;
      });
    }, 1000);

    return () => window.clearInterval(id);
  }, [activeMatchId, matchFoundVerdict]);

  async function login(nextUsername: string) {
    setBusy(true);
    setError("");
    try {
      const auth = await request<DevAuthResponse>("/auth/dev", {
        method: "POST",
        body: JSON.stringify({ username: nextUsername }),
      });
      setDevSessionToken(auth.token);
      await loadSession();
      await loadInventory();
      await loadQueueStatus();
      await loadActiveMatch();
      setUsername(nextUsername);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setBusy(false);
    }
  }

  async function loginAsUserId() {
    const parsed = Number(userId);
    if (!Number.isInteger(parsed) || parsed <= 0) {
      setError("Введите корректный user ID");
      return;
    }

    setBusy(true);
    setError("");
    try {
      const auth = await request<DevAuthResponse>("/auth/dev", {
        method: "POST",
        body: JSON.stringify({ user_id: parsed }),
      });
      setDevSessionToken(auth.token);
      await loadSession();
      await loadInventory();
      await loadQueueStatus();
      await loadActiveMatch();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setBusy(false);
    }
  }

  async function chooseHero(hero: Hero) {
    setError("");
    try {
      await request("/heroes/select", {
        method: "POST",
        body: JSON.stringify({ hero_code: hero.hero_code }),
      });
      const nextMe = await request<MeResponse>("/me");
      setSelectedHeroCode(hero.hero_code);
      setMe(nextMe);
      setScreen("menu");
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to select hero";
      setError(message);
      throw new Error(message);
    }
  }

  async function joinQueue() {
    if (!selectedHeroCode) {
      setQueueError("НЕТ ГЕРОЯ");
      return;
    }

    if (deckCardCount !== 20) {
      setQueueError("НЕПОЛНАЯ ДЕКА");
      return;
    }

    setQueueBusy(true);
    setQueueError("");

    try {
      const response = await request<JoinQueueResponse>("/queue/join", {
        method: "POST",
      });
      setQueueStatus((current) => ({
        ...current,
        state: response.state,
        opponent_user_id: response.opponent_user_id,
        search_duration_sec: 0,
      }));
      setGameModeOpen(false);
      await loadQueueStatus();
    } catch (err) {
      setQueueError(err instanceof Error ? err.message : "НЕ УДАЛОСЬ НАЧАТЬ ПОИСК");
    } finally {
      setQueueBusy(false);
    }
  }

  async function leaveQueue() {
    setQueueBusy(true);

    try {
      await request("/queue/leave", {
        method: "POST",
      });
      setQueueStatus({ state: "idle", search_duration_sec: 0 });
      setQueueError("");
    } catch (err) {
      setQueueError(err instanceof Error ? err.message : "НЕ УДАЛОСЬ ОТМЕНИТЬ ПОИСК");
    } finally {
      setQueueBusy(false);
    }
  }

  async function acceptFoundMatch() {
    setQueueBusy(true);
    setQueueError("");
    try {
      const response = await request<QueueStatusResponse | MaskedBattleMatchState>("/queue/accept", {
        method: "POST",
      });
      if ("match_id" in response) {
        setActiveMatchId(response.match_id);
        setMatchFoundVerdict("countdown");
        return;
      }
      setQueueStatus((current) => ({
        ...current,
        state: response.state,
      }));
      setAcceptedByMeLatch(true);
      setMatchFoundVerdict("accepted");
      await loadQueueStatus();
    } catch (err) {
      setQueueError(err instanceof Error ? err.message : "НЕ УДАЛОСЬ ПРИНЯТЬ МАТЧ");
    } finally {
      setQueueBusy(false);
    }
  }

  async function declineFoundMatch() {
    setQueueBusy(true);
    setQueueError("");
    try {
      await request("/queue/decline", {
        method: "POST",
      });
      setMatchFoundVerdict("declined_self");
      setQueueStatus((current) => ({ ...current, state: "penalty" }));
      window.setTimeout(() => {
        setScreen("menu");
        setAcceptedByMeLatch(false);
      }, 3000);
    } catch (err) {
      setQueueError(err instanceof Error ? err.message : "НЕ УДАЛОСЬ ОТКАЗАТЬСЯ");
    } finally {
      setQueueBusy(false);
    }
  }

  async function toggleFullscreen() {
    setError("");
    const webApp = (window as TelegramWebAppWindow).Telegram?.WebApp;
    const isActive = Boolean(document.fullscreenElement || webApp?.isFullscreen || fullscreenActive);

    try {
      if (isActive) {
        webApp?.exitFullscreen?.();
        if (document.fullscreenElement) {
          await document.exitFullscreen();
        }
        setFullscreenActive(false);
        return;
      }

      webApp?.expand?.();
      if (webApp?.requestFullscreen) {
        webApp.requestFullscreen();
        setFullscreenActive(true);
        return;
      }

      await document.documentElement.requestFullscreen();
    } catch (err) {
      webApp?.expand?.();
      setFullscreenActive(Boolean(document.fullscreenElement || webApp?.isFullscreen));
      setError(err instanceof Error ? err.message : "Fullscreen is unavailable");
    }
  }

  return (
    <main className="app-shell">
      {screen === "menu" ? (
        <MainMenu
          me={me}
          selectedHero={selectedHero}
          heroes={heroes}
          heroPickerOpen={heroPickerOpen}
          setHeroPickerOpen={setHeroPickerOpen}
          chooseHero={chooseHero}
          onStartMatch={() => {
            if (!canJoinQueue) {
              setQueueError(queueHint || "РќР•Р›Р¬Р—РЇ РќРђР§РђРўР¬ РџРћРРЎРљ");
              setGameModeOpen(true);
              return;
            }
            setQueueError("");
            setGameModeOpen(true);
          }}
          inventoryHidden={queueSearching}
          startMatchDisabled={queueSearching || !canJoinQueue}
          startMatchLabel={startMatchLabel}
          onInventory={() => setScreen("inventory")}
          musicEnabled={musicEnabled}
          onToggleMusic={() => {
            const nextEnabled = !musicEnabled;
            setMusicEnabled(nextEnabled);
            syncMusicPlayback(screenRef.current, nextEnabled, queueStatus.state);
          }}
          fullscreenActive={fullscreenActive}
          onToggleFullscreen={() => {
            void toggleFullscreen();
          }}
        />
      ) : screen === "inventory" ? (
        <InventoryScreen
          draftDeckEntries={draftDeckEntries}
          savedDeckEntries={deckEntries}
          battleCards={battleCards}
          buffCards={buffCards}
          onBack={() => setScreen("menu")}
          onDraftDeckChange={setDraftDeckEntries}
          onDeckSaved={setDeckEntries}
        />
      ) : activeMatchId && me && !matchTransitionLocked ? (
        <BattleScreen
          currentUserId={me.user_id}
          matchId={activeMatchId}
          heroes={heroes}
          battleCards={battleCards}
          deckEntries={deckEntries}
          onLeaveToMenu={() => {
            setActiveMatchId(null);
            setQueueStatus({ state: "idle", search_duration_sec: 0 });
            setQueueError("");
            setMatchFoundVerdict("idle");
            setMatchFoundCountdown(3);
            setAcceptedByMeLatch(false);
            setScreen("menu");
            void loadSession().catch(() => undefined);
            void loadQueueStatus().catch(() => undefined);
          }}
        />
      ) : null}

      {false ? (
      <section className="card surface dev-panel">
        <div className="section-head">
          <div>
            <p className="eyebrow">Dev Auth</p>
            <h2>Быстрый вход</h2>
          </div>
          <span className={`status-pill ${me ? "status-pill--ok" : "status-pill--idle"}`}>
            {busy ? "Loading" : me ? "Authorized" : "Guest"}
          </span>
        </div>

        <label className="field">
          <span>Username</span>
          <input
            value={username}
            onChange={(event) => setUsername(event.target.value)}
            placeholder="dev"
            autoComplete="off"
          />
        </label>

        <label className="field">
          <span>User ID</span>
          <input
            value={userId}
            onChange={(event) => setUserId(event.target.value)}
            placeholder="4"
            inputMode="numeric"
            autoComplete="off"
          />
        </label>

        <div className="quick-row">
          {QUICK_USERS.map((item) => (
            <button key={item} type="button" className="chip" onClick={() => login(item)} disabled={busy}>
              {item}
            </button>
          ))}
        </div>

        <button type="button" className="primary-button" onClick={() => login(username)} disabled={busy}>
          {busy ? "Подключаем..." : "Войти как dev user"}
        </button>

        <button type="button" className="secondary-button" onClick={loginAsUserId} disabled={busy}>
          {busy ? "Подключаем..." : "Войти по user ID"}
        </button>

        {error ? <p className="error-text">{error}</p> : null}
      </section>
      ) : !me && error ? (
        <section className="card surface dev-panel">
          <p className="error-text">{error}</p>
        </section>
      ) : null}

      <GameModePanel
        open={gameModeOpen}
        searching={queueSearching}
        queueState={queueStatus.state}
        penaltyUntil={queueStatus.penalty_until}
        busy={queueBusy}
        error={queueError}
        queueHint={queueHint}
        searchDurationSec={queueStatus.search_duration_sec ?? 0}
        canQueue={canJoinQueue}
        onClose={() => {
          setGameModeOpen(false);
          setQueueError("");
        }}
        onFindMatch={joinQueue}
        onCancelSearch={leaveQueue}
      />

      <MatchFoundPanel
        open={
          queueStatus.state === "pending_match" ||
          matchFoundVerdict === "declined_self" ||
          matchFoundVerdict === "declined_opponent" ||
          matchFoundVerdict === "countdown" ||
          (Boolean(activeMatchId) && acceptedByMeLatch)
        }
        busy={queueBusy}
        error={queueError}
        acceptedByMe={Boolean(queueStatus.accepted_by_me) || acceptedByMeLatch}
        acceptedByOpponent={Boolean(queueStatus.accepted_by_opponent)}
        deadlineSec={Math.max(0, Math.ceil(((queueStatus.accept_deadline_at ? Date.parse(queueStatus.accept_deadline_at) : 0) - Date.now()) / 1000))}
        verdict={matchFoundVerdict}
        countdownSec={matchFoundCountdown}
        playerRating={me?.rating}
        opponentRating={queueStatus.opponent_rating ?? 0}
        onAccept={acceptFoundMatch}
        onDecline={declineFoundMatch}
      />
    </main>
  );
}
