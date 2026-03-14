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

async function loadChatHistory() {
  const username = getStoredUsername();
  if (!username) return;

  const list = document.getElementById('chat-history-list');
  if (!list) return;

  list.innerHTML = '<div class="chat-history-empty">Loading...</div>';

  try {
    const r = await fetch('/v1/chat/history/' + encodeURIComponent(username), {
      method: 'GET',
      headers: {
        'Authorization': 'Bearer ' + getAccessToken(),
        'X-Refresh-Token': getRefreshToken()
      }
    });

    const newAccess = r.headers.get('X-New-Access-Token');
    if (newAccess) saveTokens(newAccess, getRefreshToken());

    const body = await r.json().catch(() => ({}));
    const chats = body.chats || (body.body && body.body.chats) || [];

    if (!r.ok) {
      list.innerHTML = '<div class="chat-history-empty">Could not load chat history.</div>';
      return;
    }

    renderChatHistory(chats);
  } catch (e) {
    console.error('loadChatHistory error', e);
    list.innerHTML = '<div class="chat-history-empty">Network error while loading chat history.</div>';
  }
}

function renderChatHistory(chats) {
  const list = document.getElementById('chat-history-list');
  if (!list) return;

  list.innerHTML = '';

  if (!Array.isArray(chats) || chats.length === 0) {
    list.innerHTML = '<div class="chat-history-empty">No chats yet.</div>';
    return;
  }

  for (const chat of chats) {
    const chatId = chat.chatId || chat.chat_id || '';
    const peer = chat.peerUsername || chat.peer_username || 'Unknown';
    const lastMessage = chat.lastMessage || chat.last_message || 'Нет сообщений';
    const hasUnread = !!(chat.hasUnread || chat.has_unread);
    const matchHint = chat.matchHint || chat.match_hint || '';

    const item = document.createElement('button');
    item.type = 'button';
    item.className = 'chat-history-item';

    const avatar = document.createElement('span');
    avatar.className = 'chat-history-avatar';
    avatar.textContent = peer.charAt(0).toUpperCase();

    const body = document.createElement('span');
    body.className = 'chat-history-body';

    const nameEl = document.createElement('span');
    nameEl.className = 'chat-history-name';
    nameEl.textContent = peer;

    const lastEl = document.createElement('span');
    lastEl.className = 'chat-history-last';
    lastEl.textContent = lastMessage;

    body.appendChild(nameEl);
    body.appendChild(lastEl);
    item.appendChild(avatar);
    item.appendChild(body);

    if (hasUnread) {
      const unreadDot = document.createElement('span');
      unreadDot.className = 'chat-history-unread';
      unreadDot.setAttribute('aria-label', 'Unread message');
      item.appendChild(unreadDot);
    }

    item.onclick = () => {
      const url = '/static/html/chat.html?chat_id=' + encodeURIComponent(chatId)
        + '&peer=' + encodeURIComponent(peer)
        + '&match_hint=' + encodeURIComponent(matchHint);
      window.location.href = url;
    };

    list.appendChild(item);
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

  clearFieldInvalid(document.getElementById('upd-age'));

  document.querySelectorAll('input[name="upd-interest"]').forEach(el => { el.checked = false; });
  document.querySelectorAll('input[name="edit-lang"]').forEach(el => { el.checked = false; });
  document.querySelectorAll('select[name="edit-lang-level-en"], select[name="edit-lang-level-ru"]').forEach(el => {
    el.value = '';
    clearFieldInvalid(el);
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

function toggleSection(contentId, btnId) {
  const content = document.getElementById(contentId);
  const btn = document.getElementById(btnId);
  const hidden = content.style.display === 'none';

  content.style.display = hidden ? 'flex' : 'none';
  btn.textContent = hidden ? 'Hide' : 'Show';
}

async function handleUpdateProfile() {
  const first_name = document.getElementById('upd-first-name').value.trim();
  const last_name = document.getElementById('upd-last-name').value.trim();
  const ageField = document.getElementById('upd-age');
  const ageRaw = ageField.value.trim();
  const age = parseInt(ageRaw, 10);

  const languages = collectLanguages('edit');
  const checkedLangs = document.querySelectorAll('input[name="edit-lang"]:checked').length;
  const interests = collectCheckedValues('upd-interest');

  if (!ageRaw || Number.isNaN(age) || age < 0) {
    markFieldInvalid(ageField);
    showMessage('update-result', 'Please enter a valid age.', 'error');
    return;
  }
  clearFieldInvalid(ageField);

  if (languages.length !== checkedLangs) {
    showMessage('update-result', 'Choose a level for each selected language.', 'error');
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
}

function bindProfileEvents() {
  document.getElementById('btn-profile').onclick = async () => {
    showPanel('profile');
    await loadProfile();
  };

  document.getElementById('btn-edit-profile').onclick = () => {
    const form = document.getElementById('edit-profile-form');
    form.style.display = form.style.display === 'none' ? 'block' : 'none';
  };

  document.getElementById('do-update-profile').onclick = handleUpdateProfile;

  document.getElementById('btn-chat-history').onclick = async () => {
    const block = document.getElementById('chat-history-block');
    const isHidden = block.style.display === 'none';

    block.style.display = isHidden ? 'block' : 'none';

    if (isHidden) {
      await loadChatHistory();
    }
  };

  document.getElementById('toggle-interests').onclick = () => {
    toggleSection('profile-interests', 'toggle-interests');
  };

  document.getElementById('toggle-languages').onclick = () => {
    toggleSection('profile-languages', 'toggle-languages');
  };
}
