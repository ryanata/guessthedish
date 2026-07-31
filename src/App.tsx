import { useEffect, useState } from "react";
import { createMatch, createRoom, deleteMatch, getCatalog, getMatch, joinRoom, submitGuess } from "./api";
import { GameScreen } from "./screens/GameScreen";
import { HomeScreen } from "./screens/HomeScreen";
import { JoinRoomScreen } from "./screens/JoinRoomScreen";
import { OpponentLeftScreen } from "./screens/OpponentLeftScreen";
import { RoomScreen } from "./screens/RoomScreen";
import { SearchingScreen } from "./screens/SearchingScreen";
import type { Dish, MatchSnapshot, Screen } from "./types";

const guestNames = ["QuickWhisk", "PepperMill", "HotPlate", "SousChef", "TableSeven"];

function resolvedGuestName(name: string) {
  return name.trim() || guestNames[Math.floor(Math.random() * guestNames.length)];
}

export function App() {
  const initialRoomCode = window.location.pathname.match(/^\/room\/([^/]+)\/?$/)?.[1];
  const [screen, setScreen] = useState<Screen>(initialRoomCode ? "join-room" : "home");
  const [guestName, setGuestName] = useState("");
  const [activeGuestName, setActiveGuestName] = useState("");
  const [dishes, setDishes] = useState<Dish[]>([]);
  const [match, setMatch] = useState<MatchSnapshot>();
  const [matchToken, setMatchToken] = useState("");
  const [error, setError] = useState("");
  const matchID = match?.id;
  const matchPhase = match?.phase;

  useEffect(() => {
    getCatalog()
      .then((catalog) => {
        setDishes(catalog.dishes);
        setError("");
      })
      .catch((reason: Error) => setError(reason.message));
  }, []);

  useEffect(() => {
    if (!matchID || !matchToken) return;
    let active = true;
    const timer = window.setInterval(() => {
      getMatch(matchID, matchToken)
        .then((snapshot) => {
          if (active) {
            setMatch(snapshot);
            if (screen === "room" && snapshot.phase !== "waiting") setScreen("searching");
            setError("");
          }
        })
        .catch((reason: Error) => {
          if (!active) return;
          if (reason.message === "match not found") {
            setMatch(undefined);
            setMatchToken("");
            setError("");
            setScreen("opponent-left");
            return;
          }
          setError(reason.message);
        });
    }, 250);
    return () => {
      active = false;
      window.clearInterval(timer);
    };
  }, [matchID, matchToken, screen]);

  useEffect(() => {
    if (screen !== "searching" || !matchID || matchPhase === "waiting") return;
    const timer = window.setTimeout(() => setScreen("game"), 3000);
    return () => window.clearTimeout(timer);
  }, [matchID, matchPhase, screen]);

  async function startQuickPlay() {
    const name = resolvedGuestName(guestName);
    setActiveGuestName(name);
    setError("");
    setMatch(undefined);
    setMatchToken("");
    setScreen("searching");
    try {
      const joined = await createMatch(name);
      setMatchToken(joined.token ?? "");
      setMatch(joined);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Could not start match");
      setScreen("home");
    }
  }

  async function startRoom() {
    const name = resolvedGuestName(guestName);
    setActiveGuestName(name);
    setError("");
    try {
      const created = await createRoom(name);
      setMatchToken(created.token ?? "");
      setMatch(created);
      window.history.replaceState(null, "", `/room/${created.roomCode}`);
      setScreen("room");
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Could not create room");
    }
  }

  async function acceptRoomInvite() {
    if (!initialRoomCode) return;
    const name = resolvedGuestName(guestName);
    setActiveGuestName(name);
    setError("");
    try {
      const joined = await joinRoom(initialRoomCode, name);
      setMatchToken(joined.token ?? "");
      setMatch(joined);
      setScreen("searching");
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Could not join room");
    }
  }

  function returnHome() {
    window.history.replaceState(null, "", "/");
    setError("");
    setScreen("home");
  }

  async function leaveMatch() {
    const activeMatch = match;
    const activeToken = matchToken;
    setMatch(undefined);
    setMatchToken("");
    setError("");
    window.history.replaceState(null, "", "/");
    setScreen("home");
    if (activeMatch && activeToken) await deleteMatch(activeMatch.id, activeToken).catch(() => undefined);
  }

  async function playAgain() {
    const activeMatch = match;
    const activeToken = matchToken;
    setError("");
    setMatch(undefined);
    setMatchToken("");
    setScreen("searching");
    try {
      const nextMatch = await createMatch(activeGuestName || match?.player.name || resolvedGuestName(guestName));
      setMatchToken(nextMatch.token ?? "");
      setMatch(nextMatch);
      if (activeMatch && activeToken) await deleteMatch(activeMatch.id, activeToken).catch(() => undefined);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Could not start match");
      setMatch(undefined);
      setScreen("home");
    }
  }

  async function guess(dish: Dish) {
    if (!match) return;
    try {
      setMatch(await submitGuess(match.id, matchToken, dish.id));
      setError("");
    } catch (reason) {
      if (reason instanceof Error && ![
        "player is temporarily locked",
        "match is not accepting guesses",
      ].includes(reason.message)) {
        setError(reason.message);
      }
    }
  }

  if (screen === "searching") {
    return <SearchingScreen match={match} onCancel={leaveMatch} />;
  }

  if (screen === "opponent-left") {
    return <OpponentLeftScreen onLeave={returnHome} />;
  }

  if (screen === "join-room") {
    return <JoinRoomScreen guestName={guestName} error={error} onGuestNameChange={setGuestName} onJoin={acceptRoomInvite} onLeave={returnHome} />;
  }

  if (screen === "room") {
    return (
      <RoomScreen
        match={match!}
        onLeave={leaveMatch}
      />
    );
  }

  if (screen === "game" && match) {
    return <GameScreen dishes={dishes} match={match} error={error} onGuess={guess} onLeave={leaveMatch} onPlayAgain={playAgain} />;
  }

  return (
    <HomeScreen
      guestName={guestName}
      onGuestNameChange={setGuestName}
      error={error}
      onQuickPlay={startQuickPlay}
      onCreateRoom={startRoom}
    />
  );
}
