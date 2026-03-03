// ===== PANELS =====
document.getElementById('btn-register').onclick = () => showPanel('register');
document.getElementById('btn-login').onclick = () => showPanel('login');
document.getElementById('btn-profile').onclick = async () => {
  showPanel('profile');
  await loadProfile();
};
document.getElementById('btn-matching').onclick = () => showPanel('matching');

document.getElementById('btn-edit-profile').onclick = () => {
  const form = document.getElementById('edit-profile-form');
  form.style.display = form.style.display === 'none' ? 'block' : 'none';
};

function showPanel(name) {
  for (const id of ['register', 'login', 'profile', 'matching']) {
    const el = document.getElementById(`panel-${id}`);
    if (el) el.style.display = id === name ? 'block' : 'none';
  }
}

// ===== STORAGE =====
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
function getAccessToken() { return localStorage.getItem(STORAGE_ACCESS); }
function getRefreshToken() { return localStorage.getItem(STORAGE_REFRESH); }

function saveUsername(username) {
  if (username) localStorage.setItem(STORAGE_USERNAME, username);
}
function getStoredUsername() {
  return localStorage.getItem(STORAGE_USERNAME);
}
function clearUsername() {
  localStorage.removeItem(STORAGE_USERNAME);
}

// ===== AUTH UI =====
function setAuthorizedUI(isAuthorized) {
  document.getElementById('btn-register').style.display = isAuthorized ? 'none' : 'inline-flex';
  document.getElementById('btn-login').style.display = isAuthorized ? 'none' : 'inline-flex';
  document.getElementById('btn-profile').style.display = isAuthorized ? 'inline-flex' : 'none';
  document.getElementById('btn-matching').style.display = isAuthorized ? 'inline-flex' : 'none';
  document.getElementById('btn-logout-top').style.display = isAuthorized ? 'inline-flex' : 'none';

  if (!isAuthorized) {
    showPanel('login');
  }
}

document.getElementById('btn-logout-top').onclick = logout;

function logout() {
  clearTokens();
  clearUsername();

  stopQueuePolling();
  stopQueueVisuals();
  wasInQueue = false;

  closeEditProfileForm();
  resetLoginForm();
  resetRegisterForm();

  setAuthorizedUI(false);
  showPanel('login');
}

// ===== HELPERS =====
function collectCheckedValues(name) {
  const els = document.querySelectorAll(`input[name="${name}"]:checked`);
  return Array.from(els).map(el => ({ name: el.value }));
}

function collectLanguages(prefix) {
  const els = document.querySelectorAll(`input[name="${prefix}-lang"]:checked`);
  const langs = [];
  for (const el of els) {
    const select = document.querySelector(`select[name="${prefix}-lang-level-${el.value}"]`);
    if (select && select.value) {
      langs.push({ name: el.value, level: parseInt(select.value, 10) });
    }
  }
  return langs;
}

function capitalize(v) {
  if (!v) return '';
  return v.charAt(0).toUpperCase() + v.slice(1);
}

function showMessage(boxId, text, type = 'info') {
  const el = document.getElementById(boxId);
  if (!el) return;
  el.textContent = text;
  el.className = `result ui-message ${type}`;
}

function clearMessage(boxId) {
  const el = document.getElementById(boxId);
  if (!el) return;
  el.textContent = '';
  el.className = 'result';
}

function resetLoginForm() {
  document.getElementById('login-username').value = '';
  document.getElementById('login-password').value = '';
  clearMessage('login-result');
}

function resetRegisterForm() {
  document.getElementById('reg-username').value = '';
  document.getElementById('reg-first-name').value = '';
  document.getElementById('reg-last-name').value = '';
  document.getElementById('reg-email').value = '';
  document.getElementById('reg-password').value = '';
  document.getElementById('reg-age').value = '';
  document.getElementById('reg-gender').value = 'male';

  document.querySelectorAll('input[name="reg-lang"]').forEach(el => el.checked = false);
  document.querySelectorAll('select[name="reg-lang-level-en"], select[name="reg-lang-level-ru"]').forEach(el => el.value = '');
  document.querySelectorAll('input[name="interest"]').forEach(el => el.checked = false);

  clearMessage('reg-result');
}

function closeEditProfileForm() {
  const form = document.getElementById('edit-profile-form');
  if (form) form.style.display = 'none';
  clearMessage('update-result');
}

function normalizeLangLevelToSelectValue(level) {
  if (level == null) return '';
  const raw = String(level).trim().toUpperCase();

  if (raw === '1' || raw === 'NATIVE') return '1';
  if (raw === '2' || raw === 'MEDIUM') return '2';
  if (raw === '3' || raw === 'LOW') return '3';

  return '';
}

function friendlyRegisterError(message) {
  const msg = String(message || '').toLowerCase();

  if (msg.includes('users_pkey') || msg.includes('duplicate key value')) {
    return 'This username already exists.';
  }
  if (msg.includes('users_email_key')) {
    return 'This email is already in use.';
  }
  return 'Unable to create account. Please try again.';
}

// ===== REGISTER =====
document.getElementById('do-register').onclick = async () => {
  const username = document.getElementById('reg-username').value.trim();
  const first_name = document.getElementById('reg-first-name').value.trim();
  const last_name = document.getElementById('reg-last-name').value.trim();
  const email = document.getElementById('reg-email').value.trim();
  const password = document.getElementById('reg-password').value.trim();
  const ageRaw = document.getElementById('reg-age').value.trim();
  const age = parseInt(ageRaw, 10);
  const gender = document.getElementById('reg-gender').value;

  const languages = collectLanguages('reg');
  const interests = collectCheckedValues('interest');

  if (!username || !first_name || !last_name || !email || !password || !ageRaw) {
    showMessage('reg-result', 'Please fill in all required fields.', 'error');
    return;
  }

  if (Number.isNaN(age) || age < 0) {
    showMessage('reg-result', 'Please enter a valid age.', 'error');
    return;
  }

  showMessage('reg-result', 'Creating account...', 'info');

  try {
    const r = await fetch('/v1/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        user: { username, first_name, last_name, email, password, age, gender, languages, interests }
      })
    });

    const data = await r.json().catch(() => ({}));

    if (r.ok) {
      showMessage('reg-result', 'Account created. Now sign in.', 'success');
      document.getElementById('reg-password').value = '';
      document.getElementById('login-username').value = username;
      document.getElementById('login-password').value = '';
      // НЕ перекидываем автоматически
      return;
    }

    const backendMessage = data?.body?.message || data?.message || '';
    showMessage('reg-result', friendlyRegisterError(backendMessage), 'error');
  } catch (e) {
    showMessage('reg-result', 'Network error. Please try again.', 'error');
  }
};

// ===== LOGIN =====
document.getElementById('do-login').onclick = async () => {
  const username = document.getElementById('login-username').value.trim();
  const password = document.getElementById('login-password').value.trim();

  if (!username || !password) {
    showMessage('login-result', 'Enter username and password.', 'error');
    return;
  }

  showMessage('login-result', 'Signing in...', 'info');

  try {
    const r = await fetch('/v1/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    });

    const data = await r.json().catch(() => ({}));

    const access = data.accessToken || data.access_token;
    const refresh = data.refreshToken || data.refresh_token;

    if (r.ok && access) {
      saveTokens(access, refresh);
      saveUsername(username);

      // очищаем форму логина, чтобы после logout там ничего не висело
      document.getElementById('login-password').value = '';
      clearMessage('login-result');

      setAuthorizedUI(true);
      showPanel('profile');
      await loadProfile();
      return;
    }

    showMessage('login-result', 'Unable to sign in. Check username or password.', 'error');
  } catch (e) {
    showMessage('login-result', 'Network error. Please try again.', 'error');
  }
};

// ===== PROFILE =====
async function loadProfile() {
  const username = getStoredUsername();
  if (!username) return;

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
    const user = body.user || (body.body && body.body.user);

    if (!r.ok || !user) return;

    renderProfile(user);
    fillEditProfile(user);
  } catch (e) {
    console.error('loadProfile error', e);
  }
}

function renderProfile(user) {
  const firstName = user.firstName || user.first_name || '';
  const lastName = user.lastName || user.last_name || '';
  const fullName = `${firstName} ${lastName}`.trim() || user.username || 'User';

  document.getElementById('profile-display-name').textContent = fullName;
  document.getElementById('profile-display-username').textContent = '@' + (user.username || '');
  document.getElementById('profile-avatar').textContent = (user.username || 'U').charAt(0).toUpperCase();

  document.getElementById('profile-age-pill').textContent = 'Age: ' + (user.age ?? '—');
  document.getElementById('profile-gender-pill').textContent = 'Gender: ' + capitalize(user.gender || '—');

  const interestsWrap = document.getElementById('profile-interests');
  interestsWrap.innerHTML = '';
  const interests = Array.isArray(user.interests) ? user.interests : [];
  if (interests.length === 0) {
    interestsWrap.innerHTML = '<div class="meta-pill">No interests</div>';
  } else {
    for (const it of interests) {
      const div = document.createElement('div');
      div.className = 'meta-pill';
      div.textContent = capitalize(it.name || it.interest || 'Unknown');
      interestsWrap.appendChild(div);
    }
  }

  const langsWrap = document.getElementById('profile-languages');
  langsWrap.innerHTML = '';
  const langs = Array.isArray(user.languages) ? user.languages : [];
  if (langs.length === 0) {
    langsWrap.innerHTML = '<div class="meta-pill">No languages</div>';
  } else {
    for (const l of langs) {
      const div = document.createElement('div');
      div.className = 'meta-pill';
      const name = l.name || l.language || 'Unknown';
      const level = l.level ?? '';
      div.textContent = level ? `${capitalize(name)} · ${level}` : capitalize(name);
      langsWrap.appendChild(div);
    }
  }
}

function fillEditProfile(user) {
  document.getElementById('upd-first-name').value = user.firstName || user.first_name || '';
  document.getElementById('upd-last-name').value = user.lastName || user.last_name || '';
  document.getElementById('upd-age').value = user.age ?? '';

  document.querySelectorAll('input[name="upd-interest"]').forEach(el => { el.checked = false; });
  document.querySelectorAll('input[name="edit-lang"]').forEach(el => { el.checked = false; });
  document.querySelectorAll('select[name="edit-lang-level-en"], select[name="edit-lang-level-ru"]').forEach(el => {
    el.value = '';
  });

  const interests = Array.isArray(user.interests) ? user.interests : [];
  for (const it of interests) {
    const val = String(it.name || it.interest || '').toLowerCase();
    const el = document.querySelector(`input[name="upd-interest"][value="${val}"]`);
    if (el) el.checked = true;
  }

  const langs = Array.isArray(user.languages) ? user.languages : [];
  for (const l of langs) {
    const name = String(l.name || l.language || '').toLowerCase();
    const levelValue = normalizeLangLevelToSelectValue(l.level);

    const checkbox = document.querySelector(`input[name="edit-lang"][value="${name}"]`);
    const select = document.querySelector(`select[name="edit-lang-level-${name}"]`);

    if (checkbox) checkbox.checked = true;
    if (select) select.value = levelValue;
  }
}

document.getElementById('do-update-profile').onclick = async () => {
  const first_name = document.getElementById('upd-first-name').value.trim();
  const last_name = document.getElementById('upd-last-name').value.trim();
  const ageRaw = document.getElementById('upd-age').value.trim();
  const age = parseInt(ageRaw, 10);
  const interests = collectCheckedValues('upd-interest');
  const languages = collectLanguages('edit');

  if (!ageRaw || Number.isNaN(age) || age < 0) {
    showMessage('update-result', 'Please enter a valid age.', 'error');
    return;
  }

  showMessage('update-result', 'Saving changes...', 'info');

  try {
    const r = await fetch('/v1/profile', {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + getAccessToken()
      },
      body: JSON.stringify({
        user: { first_name, last_name, age, interests, languages }
      })
    });

    const data = await r.json().catch(() => ({}));

    if (r.ok) {
      await loadProfile();
      showMessage('update-result', 'Profile updated.', 'success');

      setTimeout(() => {
        closeEditProfileForm();
      }, 500);

      return;
    }

    const msg = data?.body?.message || data?.message || 'Unable to update profile.';
    showMessage('update-result', msg, 'error');
  } catch (e) {
    showMessage('update-result', 'Network error. Please try again.', 'error');
  }
};

// ===== PROFILE TOGGLES =====
document.getElementById('toggle-interests').onclick = () => {
  toggleSection('profile-interests', 'toggle-interests');
};

document.getElementById('toggle-languages').onclick = () => {
  toggleSection('profile-languages', 'toggle-languages');
};

function toggleSection(contentId, btnId) {
  const content = document.getElementById(contentId);
  const btn = document.getElementById(btnId);
  const hidden = content.style.display === 'none';
  content.style.display = hidden ? 'flex' : 'none';
  btn.textContent = hidden ? 'Hide' : 'Show';
}

// ===== MATCHING =====
let queuePollTimer = null;
let queueWaitTimer = null;
let queueStartedAt = null;
let wasInQueue = false;

function setMatchStatus(text, type = 'idle') {
  const el = document.getElementById('match-status');
  el.textContent = text;
  el.className = `match-status ${type}`;
}

function setMatchResult(obj) {
  const box = document.getElementById('match-result');
  box.textContent = typeof obj === 'string' ? obj : JSON.stringify(obj, null, 2);
}

function startQueueVisuals() {
  document.getElementById('queue-loader').style.display = 'block';
  document.getElementById('queue-meta').style.display = 'flex';

  queueStartedAt = Date.now();

  if (queueWaitTimer) clearInterval(queueWaitTimer);
  queueWaitTimer = setInterval(() => {
    const sec = Math.floor((Date.now() - queueStartedAt) / 1000);
    document.getElementById('queue-wait-time').textContent = `${sec}s`;
  }, 1000);
}

function stopQueueVisuals() {
  document.getElementById('queue-loader').style.display = 'none';
  document.getElementById('queue-meta').style.display = 'none';

  if (queueWaitTimer) {
    clearInterval(queueWaitTimer);
    queueWaitTimer = null;
  }
  queueStartedAt = null;
}

function stopQueuePolling() {
  if (queuePollTimer) {
    clearInterval(queuePollTimer);
    queuePollTimer = null;
  }
}

async function refreshQueueState(username) {
  try {
    const r = await fetch('/v1/matching/list');
    const data = await r.json();
    const list = data.usernames || data.Usernames || [];
    document.getElementById('queue-count').textContent = list.length;

    const isInQueue = list.includes(username);

    if (isInQueue) {
      wasInQueue = true;
      setMatchStatus('Searching for a partner...', 'searching');
    } else if (wasInQueue) {
      stopQueuePolling();
      stopQueueVisuals();
      wasInQueue = false;
      setMatchStatus('You may have been matched. Check your chat.', 'success');
      setMatchResult('User disappeared from queue. Possible match found.');
    } else {
      setMatchStatus('Not searching', 'idle');
    }
  } catch (e) {
    setMatchStatus('Queue polling error', 'error');
    setMatchResult('Network error while checking queue: ' + e.message);
  }
}

function startQueuePolling(username) {
  stopQueuePolling();
  refreshQueueState(username);
  queuePollTimer = setInterval(() => refreshQueueState(username), 3000);
}

document.getElementById('btn-start-search').onclick = async () => {
  const username = getStoredUsername();
  const mode = parseInt(document.getElementById('match-mode').value, 10);

  if (!username) {
    setMatchStatus('Login first', 'error');
    setMatchResult('No username found. Please login first.');
    return;
  }

  setMatchResult('You are now in the queue.');
  setMatchResult('...joining');

  try {
    const r = await fetch('/v1/matching/join', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + getAccessToken()
      },
      body: JSON.stringify({ username, mode })
    });

    const data = await r.json();
    setMatchResult({ status: r.status, body: data });

    if (r.ok && data.ok) {
      wasInQueue = true;
      setMatchStatus('Searching for a partner...', 'searching');
      startQueueVisuals();
      startQueuePolling(username);
    } else {
      setMatchStatus('Failed to join queue', 'error');
    }
  } catch (e) {
    setMatchStatus('Join request failed', 'error');
    setMatchResult('Network error: ' + e.message);
  }
};

document.getElementById('btn-leave-search').onclick = async () => {
  const username = getStoredUsername();

  if (!username) {
    setMatchStatus('No username found', 'error');
    setMatchResult('Cannot leave queue: no username');
    return;
  }

  setMatchStatus('Leaving queue...', 'idle');
  setMatchResult('...leaving');

  try {
    const r = await fetch('/v1/matching/leave', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + getAccessToken()
      },
      body: JSON.stringify({ username })
    });

    const data = await r.json();
    setMatchResult({ status: r.status, body: data });

    if (r.ok && data.ok) {
      wasInQueue = false;
      stopQueuePolling();
      stopQueueVisuals();
      setMatchStatus('You left the queue', 'idle');
      document.getElementById('queue-count').textContent = '0';
      document.getElementById('queue-wait-time').textContent = '0s';
    } else {
      setMatchStatus('Failed to leave queue', 'error');
    }
  } catch (e) {
    setMatchStatus('Leave request failed', 'error');
    setMatchResult('Network error: ' + e.message);
  }
};

// ===== INIT =====
(function init() {
  const hasAccess = !!getAccessToken();
  const hasUsername = !!getStoredUsername();

  if (hasAccess && hasUsername) {
    setAuthorizedUI(true);
    showPanel('profile');
    loadProfile();
  } else {
    setAuthorizedUI(false);
    showPanel('login');
  }
})();
