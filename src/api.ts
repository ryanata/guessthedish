import type { Catalog, MatchSnapshot } from "./types";

type APIError = {
  error?: { message?: string };
};

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, init);
  if (!response.ok) {
    const body = await response.json().catch(() => ({})) as APIError;
    throw new Error(body.error?.message ?? `Request failed (${response.status})`);
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}

export function getCatalog() {
  return request<Catalog>("/api/catalog");
}

export function createMatch(name: string) {
  return request<MatchSnapshot>("/api/matches", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name }),
  });
}

export function createRoom(name: string) {
  return request<MatchSnapshot>("/api/rooms", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name }),
  });
}

export function joinRoom(code: string, name: string) {
  return request<MatchSnapshot>(`/api/rooms/${encodeURIComponent(code)}/join`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name }),
  });
}

function authorized(token: string, init?: RequestInit): RequestInit {
  const headers = new Headers(init?.headers);
  headers.set("Authorization", `Bearer ${token}`);
  return { ...init, headers };
}

export function getMatch(id: string, token: string) {
  return request<MatchSnapshot>(`/api/matches/${id}`, authorized(token));
}

export function submitGuess(id: string, token: string, dishId: string) {
  return request<MatchSnapshot>(`/api/matches/${id}/guesses`, authorized(token, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ dishId }),
  }));
}

export function deleteMatch(id: string, token: string) {
  return request<void>(`/api/matches/${id}`, authorized(token, { method: "DELETE" }));
}
