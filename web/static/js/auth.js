// web/static/js/auth.js

function logout() {
  clearTokens();
  clearUsername();

  if (window.AppWS) window.AppWS.close();

  if (window.AppMatching) {
    window.AppMatching.stopQueuePolling();
    window.AppMatching.stopQueueVisuals();
    window.AppMatching.resetState();
  }

  closeEditProfileForm();
  resetLoginForm();
  resetRegisterForm();

  setAuthorizedUI(false);
  showPanel('login');
}

async function performLogin(username, password, statusBoxId = 'login-result') {
  showMessage(statusBoxId, 'Signing in...', 'info');

  const r = await fetch('/v1/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password })
  });

  const data = await r.json().catch(() => ({}));
  const access = data.accessToken || data.access_token;
  const refresh = data.refreshToken || data.refresh_token;

  if (!(r.ok && access)) {
    return { ok: false };
  }

  saveTokens(access, refresh);
  saveUsername(username);

  if (window.AppWS) window.AppWS.connect(username);

  clearMessage(statusBoxId);
  setAuthorizedUI(true);
  showPanel('profile');
  await loadProfile();

  return { ok: true };
}

async function handleRegister() {
  const usernameField = document.getElementById('reg-username');
  const firstNameField = document.getElementById('reg-first-name');
  const lastNameField = document.getElementById('reg-last-name');
  const emailField = document.getElementById('reg-email');
  const passwordField = document.getElementById('reg-password');
  const ageField = document.getElementById('reg-age');

  const requiredFields = [
    usernameField, firstNameField, lastNameField, emailField, passwordField, ageField
  ];

  if (!validateRequiredFields(requiredFields)) {
    showMessage('reg-result', 'Please fill in all required fields.', 'error');
    return;
  }

  const username = usernameField.value.trim();
  const first_name = firstNameField.value.trim();
  const last_name = lastNameField.value.trim();
  const email = emailField.value.trim();
  const password = passwordField.value.trim();
  const ageRaw = ageField.value.trim();
  const age = parseInt(ageRaw, 10);
  const gender = document.getElementById('reg-gender').value;

  const languages = collectLanguages('reg');
  const checkedLangs = document.querySelectorAll('input[name="reg-lang"]:checked').length;
  const interests = collectCheckedValues('interest');

  if (Number.isNaN(age) || age < 0) {
    markFieldInvalid(ageField);
    showMessage('reg-result', 'Please enter a valid age.', 'error');
    return;
  }
  clearFieldInvalid(ageField);

  if (languages.length !== checkedLangs) {
    showMessage('reg-result', 'Choose a level for each selected language.', 'error');
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
      showMessage('reg-result', 'Account created. Signing you in...', 'info');

      const loginResult = await performLogin(username, password, 'reg-result');
      if (loginResult.ok) {
        resetRegisterForm();
        return;
      }

      showMessage('reg-result', 'Account created, but auto sign-in failed. Please log in manually.', 'error');
      document.getElementById('login-username').value = username;
      document.getElementById('login-password').value = '';
      passwordField.value = '';
      return;
    }

    const backendMessage = data?.body?.message || data?.message || '';
    showMessage('reg-result', friendlyRegisterError(backendMessage), 'error');
  } catch (e) {
    showMessage('reg-result', 'Network error. Please try again.', 'error');
  }
}

async function handleLogin() {
  const usernameField = document.getElementById('login-username');
  const passwordField = document.getElementById('login-password');

  const requiredFields = [usernameField, passwordField];

  if (!validateRequiredFields(requiredFields)) {
    showMessage('login-result', 'Enter username and password.', 'error');
    return;
  }

  const username = usernameField.value.trim();
  const password = passwordField.value.trim();

  try {
    const result = await performLogin(username, password, 'login-result');
    if (result.ok) {
      passwordField.value = '';
      return;
    }

    showMessage('login-result', 'Unable to sign in. Check username or password.', 'error');
  } catch (e) {
    showMessage('login-result', 'Network error. Please try again.', 'error');
  }
}

function bindAuthEvents() {
  document.getElementById('btn-register').onclick = () => showPanel('register');
  document.getElementById('btn-login').onclick = () => showPanel('login');
  document.getElementById('btn-logout-top').onclick = logout;

  document.getElementById('do-register').onclick = handleRegister;
  document.getElementById('do-login').onclick = handleLogin;
}

// expose if needed
window.AppAuth = { bindAuthEvents, logout };
