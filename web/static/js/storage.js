const STORAGE_ACCESS = 'accessToken';
const STORAGE_REFRESH = 'refreshToken';
const STORAGE_USERNAME = 'authUsername';
const STORAGE_MUTED_USERS = 'mutedUsers';

function saveTokens(access, refresh) {
  if (access) localStorage.setItem(STORAGE_ACCESS, access);
  if (refresh) localStorage.setItem(STORAGE_REFRESH, refresh);
}

function clearTokens() {
  localStorage.removeItem(STORAGE_ACCESS);
  localStorage.removeItem(STORAGE_REFRESH);
}

function getAccessToken() {
  return localStorage.getItem(STORAGE_ACCESS);
}

function getRefreshToken() {
  return localStorage.getItem(STORAGE_REFRESH);
}

function saveUsername(username) {
  if (username) localStorage.setItem(STORAGE_USERNAME, username);
}

function getStoredUsername() {
  return localStorage.getItem(STORAGE_USERNAME);
}

function clearUsername() {
  localStorage.removeItem(STORAGE_USERNAME);
}

function getMutedUsers() {
  try {
    const raw = localStorage.getItem(STORAGE_MUTED_USERS);
    const list = JSON.parse(raw || '[]');
    return Array.isArray(list) ? list : [];
  } catch (_) {
    return [];
  }
}

function isUserMuted(username) {
  const normalized = String(username || '').trim().toLowerCase();
  if (!normalized) return false;
  return getMutedUsers().includes(normalized);
}

function muteUser(username) {
  const normalized = String(username || '').trim().toLowerCase();
  if (!normalized) return;
  const next = Array.from(new Set([...getMutedUsers(), normalized]));
  localStorage.setItem(STORAGE_MUTED_USERS, JSON.stringify(next));
}

function unmuteUser(username) {
  const normalized = String(username || '').trim().toLowerCase();
  if (!normalized) return;
  const next = getMutedUsers().filter(item => item !== normalized);
  localStorage.setItem(STORAGE_MUTED_USERS, JSON.stringify(next));
}
