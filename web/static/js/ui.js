function showPanel(name) {
  for (const id of ['register', 'login', 'profile', 'matching', 'chat']) {
    const el = document.getElementById(`panel-${id}`);
    if (el) el.style.display = id === name ? 'block' : 'none';
  }
}

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

function capitalize(v) {
  if (!v) return '';
  return v.charAt(0).toUpperCase() + v.slice(1);
}

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
      clearFieldInvalid(select);
    } else if (select) {
      markFieldInvalid(select);
    }
  }

  return langs;
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

  if (msg.includes('users_email_key')) {
    return 'This email is already in use.';
  }
  if (msg.includes('users_pkey') || msg.includes('duplicate key value')) {
    return 'This username already exists.';
  }

  return 'Unable to create account. Please try again.';
}

function resetLoginForm() {
  document.getElementById('login-username').value = '';
  document.getElementById('login-password').value = '';

  clearFieldInvalid(document.getElementById('login-username'));
  clearFieldInvalid(document.getElementById('login-password'));
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

  const registerFields = [
    document.getElementById('reg-username'),
    document.getElementById('reg-first-name'),
    document.getElementById('reg-last-name'),
    document.getElementById('reg-email'),
    document.getElementById('reg-password'),
    document.getElementById('reg-age')
  ];
  registerFields.forEach(clearFieldInvalid);

  document.querySelectorAll('input[name="reg-lang"]').forEach(el => el.checked = false);
  document.querySelectorAll('select[name="reg-lang-level-en"], select[name="reg-lang-level-ru"]').forEach(el => {
    el.value = '';
    clearFieldInvalid(el);
  });
  document.querySelectorAll('input[name="interest"]').forEach(el => el.checked = false);

  clearMessage('reg-result');
}

function closeEditProfileForm() {
  const form = document.getElementById('edit-profile-form');
  if (form) form.style.display = 'none';
  clearMessage('update-result');
}

function markFieldInvalid(field) {
  if (!field) return;
  field.classList.add('input-error');
}

function clearFieldInvalid(field) {
  if (!field) return;
  field.classList.remove('input-error');
}

function validateRequiredFields(fields) {
  let ok = true;

  for (const field of fields) {
    const value = (field.value ?? '').toString().trim();
    if (!value) {
      markFieldInvalid(field);
      ok = false;
    } else {
      clearFieldInvalid(field);
    }
  }

  return ok;
}

function attachValidationListeners() {
  const selectors = [
    '#reg-username',
    '#reg-first-name',
    '#reg-last-name',
    '#reg-email',
    '#reg-password',
    '#reg-age',
    '#login-username',
    '#login-password',
    '#upd-age',
    'select[name="reg-lang-level-en"]',
    'select[name="reg-lang-level-ru"]',
    'select[name="edit-lang-level-en"]',
    'select[name="edit-lang-level-ru"]'
  ];

  selectors.forEach(selector => {
    const elements = document.querySelectorAll(selector);
    elements.forEach(el => {
      el.addEventListener('input', () => clearFieldInvalid(el));
      el.addEventListener('change', () => clearFieldInvalid(el));
    });
  });
}

function bindTabs() {
  const btnRegister = document.getElementById('btn-register');
  const btnLogin = document.getElementById('btn-login');
  const btnProfile = document.getElementById('btn-profile');
  const btnMatching = document.getElementById('btn-matching');

  if (btnRegister) btnRegister.onclick = () => showPanel('register');
  if (btnLogin) btnLogin.onclick = () => showPanel('login');
  if (btnProfile) btnProfile.onclick = () => showPanel('profile');
  if (btnMatching) btnMatching.onclick = () => showPanel('matching');
}