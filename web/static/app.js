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
function collectCheckedLanguages() {
  const els = document.querySelectorAll('input[name="lang"]:checked');
  if (els.length === 0) {
    alert("Please select at least one language.");
    throw new Error("No language selected");
  }

  const langs = [];

  for (const el of els) {
    const levelSelect = document.querySelector(`select[name="lang_level_${el.value}"]`);
    if (!levelSelect) {
      alert(`No level select found for ${el.value}`);
      throw new Error("Level select missing");
    }

    const levelValue = levelSelect.value;
    if (levelValue === "") {
      alert(`Please select a level for ${el.value.toUpperCase()}.`);
      throw new Error("Language level not selected");
    }

    langs.push({ name: el.value, level: parseInt(levelValue) });
  }

  return langs;
}



function collectCheckedValues(name) {
  const els = document.querySelectorAll(`input[name="${name}"]:checked`);
  return Array.from(els).map(el => ({ name: el.value }));
}

// Register
document.getElementById('do-register').onclick = async () => {
  const username = document.getElementById('reg-username').value.trim();
  const first_name = document.getElementById('reg-first-name').value.trim();
  const last_name = document.getElementById('reg-last-name').value.trim();
  const email = document.getElementById('reg-email').value.trim();
  const password = document.getElementById('reg-password').value.trim();
  const age = parseInt(document.getElementById('reg-age').value) || 0;
  const gender = document.getElementById('reg-gender').value;

  const languages = collectCheckedLanguages();
  const interests = collectCheckedValues('interest');

  // === Проверка обязательных полей ===
  if (!username || !first_name || !last_name || !email || !password || !gender) {
    alert("Please fill in all required fields (username, first_name, last_name, email, password, gender).");
    return;
  }

  if (languages.length === 0) {
    alert("Please select at least one language and its level.");
    return;
  }

  if (interests.length === 0) {
    alert("Please select at least one interest.");
    return;
  }

  const body = {
    user: {
      username,
      first_name,
      last_name,
      email,
      password,
      age,
      gender,
      languages,
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
    const r = await fetch('/v1/profile', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + getAccessToken(),
        'X-Refresh-Token': getRefreshToken()
      },
      body: JSON.stringify({})
    });

    const newAccess = r.headers.get('X-New-Access-Token');
    if (newAccess) saveTokens(newAccess, refresh);

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
