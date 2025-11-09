// Простая SPA-логика: переключение панелей
document.getElementById('btn-register').onclick = () => showPanel('register');
document.getElementById('btn-login').onclick = () => showPanel('login');
document.getElementById('btn-profile').onclick = () => showPanel('profile');

function showPanel(name) {
  document.getElementById('panel-register').style.display = name === 'register' ? 'block' : 'none';
  document.getElementById('panel-login').style.display = name === 'login' ? 'block' : 'none';
  document.getElementById('panel-profile').style.display = name === 'profile' ? 'block' : 'none';
}

// Helpers for tokens
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

function getAccessToken() {
  return localStorage.getItem(STORAGE_ACCESS);
}
function getRefreshToken() {
  return localStorage.getItem(STORAGE_REFRESH);
}

// Build registration payload from checked inputs
function collectCheckedValues(name) {
  const els = document.querySelectorAll(`input[name="${name}"]:checked`);
  return Array.from(els).map(el => ({ name: el.value }));
}

// Register
document.getElementById('do-register').onclick = async () => {
  const username = document.getElementById('reg-username').value;
  const password = document.getElementById('reg-password').value;
  const language = collectCheckedValues('lang');      // array of {name: "en"}
  const interests = collectCheckedValues('interest'); // array of {name:"music"}

  const body = {
    user: {
      username,
      password,
      language,
      interests
    }
  };

  const resBox = document.getElementById('reg-result');
  resBox.textContent = '...loading';
  try {
    const r = await fetch('/v1/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });
    const data = await r.json();
    resBox.textContent = JSON.stringify({ status: r.status, body: data }, null, 2);
  } catch (e) {
    resBox.textContent = 'Network error: ' + e.message;
  }
};

// Login
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

    // Может твой backend возвращает ключи с другими именами (accessToken vs access_token).
    // Поддержим оба варианта:
    const access = data.accessToken || data.access_token;
    const refresh = data.refreshToken || data.refresh_token;
    if (access) saveTokens(access, refresh);
  } catch (e) {
    resBox.textContent = 'Network error: ' + e.message;
  }
};

// Get profile
document.getElementById('btn-get-profile').onclick = async () => {
  const resBox = document.getElementById('profile-result');
  resBox.textContent = '...loading';
  const access = getAccessToken();
  const refresh = getRefreshToken();

  try {
    const r = await fetch('/profile', {
      method: 'GET',
      headers: {
        'Authorization': 'Bearer ' + (access || ''),
        'X-Refresh-Token': refresh || ''
      }
    });

    // If server returned new access token in headers, pick it up
    const newAccess = r.headers.get('X-New-Access-Token');
    if (newAccess) {
      saveTokens(newAccess, refresh);
    }

    const body = await r.json();
    resBox.textContent = JSON.stringify({ status: r.status, body }, null, 2);
  } catch (e) {
    resBox.textContent = 'Network error: ' + e.message;
  }
};

// Refresh manual
document.getElementById('btn-refresh-token').onclick = async () => {
  const resBox = document.getElementById('profile-result');
  resBox.textContent = '...refreshing';
  const refresh = getRefreshToken();
  if (!refresh) {
    resBox.textContent = 'no refresh token stored';
    return;
  }

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

// Logout
document.getElementById('btn-logout').onclick = () => {
  clearTokens();
  document.getElementById('profile-result').textContent = 'Logged out';
};
