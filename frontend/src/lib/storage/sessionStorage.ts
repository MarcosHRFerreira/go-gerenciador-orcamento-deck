import type { AuthSession } from "../../features/auth/types/auth";

let currentSession: AuthSession | null = null;

export function getStoredSession(): AuthSession | null {
  return currentSession;
}

export function setStoredSession(session: AuthSession) {
  currentSession = session;
}

export function clearStoredSession() {
  currentSession = null;
}
