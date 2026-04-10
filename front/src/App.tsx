import { useEffect, useState } from "react";

type MeResponse = {
  user_id: number;
  username: string;
  first_name: string;
  rating: number;
  xp: number;
  selected_hero_code?: string;
  selected_hero_name?: string;
};

const QUICK_USERS = ["dev", "roman", "test", "rosa"];

type Hero = {
  hero_id: number;
  hero_code: string;
  name: string;
  description: string;
  image_key: string;
  level: number;
  health_points: number;
  attack_power: number;
  attack_cooldown: number;
};

type HeroesResponse = {
  heroes: Hero[];
};

function resolveImageSrc(key?: string): string {
  if (!key) {
    return "/assets/placeholders/hero_image.svg";
  }
  return `/assets/${key.replace(/^\/+/, "").replace(/\/+/g, "/")}.png`;
}

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
    ...init,
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `Request failed: ${response.status}`);
  }

  return response.json() as Promise<T>;
}

export function App() {
  const [username, setUsername] = useState("dev");
  const [userId, setUserId] = useState("");
  const [me, setMe] = useState<MeResponse | null>(null);
  const [heroes, setHeroes] = useState<Hero[]>([]);
  const [busy, setBusy] = useState(true);
  const [error, setError] = useState("");
  const [selectedHeroCode, setSelectedHeroCode] = useState("");
  const [heroPickerOpen, setHeroPickerOpen] = useState(false);

  const selectedHero = heroes.find((hero) => hero.hero_code === selectedHeroCode) ?? heroes[0] ?? null;

  async function loadSession() {
    const [nextMe, nextHeroes] = await Promise.all([
      request<MeResponse>("/me"),
      request<HeroesResponse>("/heroes"),
    ]);
    setMe(nextMe);
    setHeroes(nextHeroes.heroes);
    setSelectedHeroCode(nextMe.selected_hero_code || nextHeroes.heroes[0]?.hero_code || "");
    setError("");
  }

  useEffect(() => {
    loadSession()
      .catch(() => {
        setMe(null);
        setHeroes([]);
        setSelectedHeroCode("");
      })
      .finally(() => setBusy(false));
  }, []);

  async function login(nextUsername: string) {
    setBusy(true);
    setError("");
    try {
      await request("/auth/dev", {
        method: "POST",
        body: JSON.stringify({ username: nextUsername }),
      });
      await loadSession();
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
      await request("/auth/dev", {
        method: "POST",
        body: JSON.stringify({ user_id: parsed }),
      });
      await loadSession();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setBusy(false);
    }
  }

  async function chooseHero(hero: Hero) {
    setSelectedHeroCode(hero.hero_code);
    try {
      await request("/heroes/select", {
        method: "POST",
        body: JSON.stringify({ hero_code: hero.hero_code }),
      });
      const nextMe = await request<MeResponse>("/me");
      setMe(nextMe);
      setHeroPickerOpen(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to select hero");
    }
  }

  return (
    <main className="app-shell">
      <section className="main-menu surface">
        <div className="video-stage" aria-hidden="true">
          <div className="video-stage__glow" />
          <div className="video-stage__label">video / key art zone</div>
        </div>

        <header className="menu-topbar">
          <button type="button" className="top-slot top-slot--left">
            Friends
          </button>
          <h1 className="menu-title">PROJECT ROSE</h1>
          <button type="button" className="top-slot top-slot--right">
            Balance
          </button>
        </header>

        <section className="hero-focus">
          <button type="button" className="hero-avatar" onClick={() => setHeroPickerOpen(true)}>
            {selectedHero ? (
              <img src={resolveImageSrc(selectedHero.image_key)} alt={selectedHero.name} />
            ) : (
              <span>Hero</span>
            )}
          </button>
          <div className="hero-nameplate">{selectedHero?.name ?? "No hero selected"}</div>
          <div className="player-tag">
            {me ? `${me.username} / rating ${me.rating}` : "guest / no session"}
          </div>
        </section>

        <section className="menu-actions">
          <button type="button" className="menu-button menu-button--primary">
            Start Match
          </button>
          <button type="button" className="menu-button">
            Inventory
          </button>
          <button type="button" className="menu-panel">
            Shop Placeholder
          </button>
        </section>
      </section>

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

      {heroPickerOpen ? (
        <div className="overlay" onClick={() => setHeroPickerOpen(false)}>
          <div className="picker surface" onClick={(event) => event.stopPropagation()}>
            <div className="section-head">
              <div>
                <p className="eyebrow">Hero Select</p>
                <h2>Выбор персонажа</h2>
              </div>
              <button type="button" className="picker-close" onClick={() => setHeroPickerOpen(false)}>
                Close
              </button>
            </div>
            <div className="hero-grid">
              {heroes.map((hero) => (
                <button
                  key={hero.hero_code}
                  type="button"
                  className={`hero-card ${hero.hero_code === selectedHeroCode ? "hero-card--active" : ""}`}
                  onClick={() => chooseHero(hero)}
                >
                  <span className="hero-card__avatar">
                    <img src={resolveImageSrc(hero.image_key)} alt={hero.name} />
                  </span>
                  <strong>{hero.name}</strong>
                </button>
              ))}
            </div>
          </div>
        </div>
      ) : null}
    </main>
  );
}
