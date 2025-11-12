// SPA panel switch
document.getElementById('btn-register').onclick = () => showPanel('register');
document.getElementById('btn-login').onclick = () => showPanel('login');
document.getElementById('btn-profile').onclick = () => showPanel('profile');
document.getElementById('btn-edit-profile').onclick = () => {
  const form = document.getElementById('edit-profile-form');
  form.style.display = form.style.display === 'none' ? 'block' : 'none';
};

function showPanel(name) {
  for (const id of ['register', 'login', 'profile', 'edit-profile']) {
    document.getElementById(`panel-${id}`).style.display = id === name ? 'block' : 'none';
  }
}

// token helpers
const STORAGE_ACCESS = 'accessToken';
const STORAGE_REFRESH = 'refreshToken';

function saveTokens(access, refresh) {
  if (access) localStorage.setItem(STORAGE_ACCESS, access);
  if (refresh) localStorage.setItem(STORAGE_REFRESH, refresh);
}
function clearTokens() {
  localStorage.removeItem(STORAGE_ACCESS);
  localStorage.removeItem(STORAGE_REFRESH);
}
function getAccessToken() { return localStorage.getItem(STORAGE_ACCESS); }
function getRefreshToken() { return localStorage.getItem(STORAGE_REFRESH); }

// collect helpers
function collectCheckedValues(name) {
  const els = document.querySelectorAll(`input[name="${name}"]:checked`);
  return Array.from(els).map(el => ({ name: el.value }));
}

function collectLanguages(prefix) {
  const els = document.querySelectorAll(`input[name="${prefix}-lang"]:checked`);
  const langs = [];
  for (const el of els) {
    const select = document.querySelector(`select[name="${prefix}-lang-level-${el.value}"]`);
    if (select && select.value) langs.push({ name: el.value, level: parseInt(select.value) });
  }
  return langs;
}

// === REGISTER ===
document.getElementById('do-register').onclick = async () => {
  const username = document.getElementById('reg-username').value.trim();
  const first_name = document.getElementById('reg-first-name').value.trim();
  const last_name = document.getElementById('reg-last-name').value.trim();
  const email = document.getElementById('reg-email').value.trim();
  const password = document.getElementById('reg-password').value.trim();
  const age = parseInt(document.getElementById('reg-age').value) || 0;
  const gender = document.getElementById('reg-gender').value;

  const languages = collectLanguages('reg');
  const interests = collectCheckedValues('interest');

  const resBox = document.getElementById('reg-result');
  resBox.textContent = '...loading';

  try {
    const r = await fetch('/v1/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ user: { username, first_name, last_name, email, password, age, gender, languages, interests } })
    });
    const data = await r.json();
    resBox.textContent = JSON.stringify({ status: r.status, body: data }, null, 2);
  } catch (e) {
    resBox.textContent = 'Network error: ' + e.message;
  }
};

// === LOGIN ===
document.getElementById('do-login').onclick = async () => {
  const username = document.getElementById('login-username').value;
  const password = document.getElementById('login-password').value;
  const resBox = document.getElementById('login-result');
  resBox.textContent = '...loading';
  try {
    const r = await fetch('/v1/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    });
    const data = await r.json();
    resBox.textContent = JSON.stringify({ status: r.status, body: data }, null, 2);

    const access = data.accessToken || data.access_token;
    const refresh = data.refreshToken || data.refresh_token;
    if (access) saveTokens(access, refresh);
  } catch (e) {
    resBox.textContent = 'Network error: ' + e.message;
  }
};

// === PROFILE ===
document.getElementById('btn-get-profile').onclick = async () => {
  const resBox = document.getElementById('profile-result');
  resBox.textContent = '...loading';
  try {
    const r = await fetch('/v1/profile', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + getAccessToken(),
        'X-Refresh-Token': getRefreshToken()
      }
    });
    const newAccess = r.headers.get('X-New-Access-Token');
    if (newAccess) saveTokens(newAccess, getRefreshToken());
    const body = await r.json();
    resBox.textContent = JSON.stringify({ status: r.status, body }, null, 2);
  } catch (e) {
    resBox.textContent = 'Network error: ' + e.message;
  }
};

// === REFRESH ===
document.getElementById('btn-refresh-token').onclick = async () => {
  const resBox = document.getElementById('profile-result');
  resBox.textContent = '...refreshing';
  const refresh = getRefreshToken();
  if (!refresh) return resBox.textContent = 'no refresh token stored';

  try {
    const r = await fetch('/v1/refresh-token', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refresh })
    });
    const data = await r.json();
    resBox.textContent = JSON.stringify({ status: r.status, body: data }, null, 2);
    const access = data.accessToken || data.access_token;
    const refreshed = data.refreshToken || data.refresh_token;
    if (access) saveTokens(access, refreshed || refresh);
  } catch (e) {
    resBox.textContent = 'Network error: ' + e.message;
  }
};

// === LOGOUT ===
document.getElementById('btn-logout').onclick = () => {
  clearTokens();
  document.getElementById('profile-result').textContent = 'Logged out';
};

// === EDIT PROFILE ===
document.getElementById('do-update-profile').onclick = async () => {
  const resBox = document.getElementById('update-result');
  resBox.textContent = '...saving';

  const first_name = document.getElementById('upd-first-name').value.trim();
  const last_name = document.getElementById('upd-last-name').value.trim();
  const age = parseInt(document.getElementById('upd-age').value) || 0;
  const interests = collectCheckedValues('upd-interest');
  const languages = collectLanguages('edit');

  try {
    const r = await fetch('/v1/profile', {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + getAccessToken()
      },
      body: JSON.stringify({ user: { first_name, last_name, age, interests, languages } })
    });
    const data = await r.json();
    resBox.textContent = JSON.stringify({ status: r.status, body: data }, null, 2);
  } catch (e) {
    resBox.textContent = 'Network error: ' + e.message;
  }
};
