import { useEffect, useState } from "react";

type MeResponse = {
  user_id: number;
  username: string;
  first_name: string;
  rating: number;
  xp: number;
  selected_hero_name?: string;
};

const QUICK_USERS = ["dev", "roman", "test", "rosa"];
const HEROES = [
  "Astra Vanguard",
  "Hex Runner",
  "Ivory Saint",
  "Rust King",
  "Noctis Bloom",
  "Signal Warden",
] as const;

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
  const [me, setMe] = useState<MeResponse | null>(null);
  const [busy, setBusy] = useState(true);
  const [error, setError] = useState("");
  const [selectedHero, setSelectedHero] = useState<(typeof HEROES)[number]>(HEROES[0]);
  const [heroPickerOpen, setHeroPickerOpen] = useState(false);

  useEffect(() => {
    request<MeResponse>("/me")
      .then((next) => {
        setMe(next);
        setError("");
      })
      .catch(() => {
        setMe(null);
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
      const nextMe = await request<MeResponse>("/me");
      setMe(nextMe);
      setUsername(nextUsername);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setBusy(false);
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
            <span>Hero</span>
          </button>
          <div className="hero-nameplate">{selectedHero}</div>
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
              {HEROES.map((hero) => (
                <button
                  key={hero}
                  type="button"
                  className={`hero-card ${hero === selectedHero ? "hero-card--active" : ""}`}
                  onClick={() => {
                    setSelectedHero(hero);
                    setHeroPickerOpen(false);
                  }}
                >
                  <span className="hero-card__avatar">Avatar</span>
                  <strong>{hero}</strong>
                </button>
              ))}
            </div>
          </div>
        </div>
      ) : null}
    </main>
  );
}
