// web/static/js/main.js

function initApp() {
  initInterestPickers();
  bindTabs();
  bindAuthEvents();
  bindProfileEvents();
  if (window.AppMatching) window.AppMatching.bindMatchingEvents();
  attachValidationListeners();

  const hasAccess = !!getAccessToken();
  const hasUsername = !!getStoredUsername();

  if (hasAccess && hasUsername) {
    setAuthorizedUI(true);
    showPanel('profile');
    loadProfile();

    // auto connect ws on reload
    if (window.AppWS) window.AppWS.connect(getStoredUsername());
  } else {
    setAuthorizedUI(false);
    showPanel('login');
  }
}

initApp();
