const STORAGE_ACCESS = 'accessToken';
const STORAGE_REFRESH = 'refreshToken';
const STORAGE_USERNAME = 'authUsername';

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