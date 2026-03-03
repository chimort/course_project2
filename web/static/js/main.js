function initApp() {
  bindAuthEvents();
  bindProfileEvents();
  bindMatchingEvents();
  attachValidationListeners();

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
}

initApp();