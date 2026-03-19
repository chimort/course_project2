const INTEREST_CATALOG = [
  { value: 'music', label: 'Music', category: 'Culture' },
  { value: 'movies', label: 'Movies', category: 'Culture' },
  { value: 'books', label: 'Books', category: 'Culture' },
  { value: 'art', label: 'Art', category: 'Culture' },
  { value: 'painting', label: 'Painting', category: 'Culture' },
  { value: 'photography', label: 'Photography', category: 'Culture' },
  { value: 'design', label: 'Design', category: 'Culture' },
  { value: 'theatre', label: 'Theatre', category: 'Culture' },
  { value: 'dance', label: 'Dance', category: 'Culture' },
  { value: 'sport', label: 'Sport', category: 'Active' },
  { value: 'fitness', label: 'Fitness', category: 'Active' },
  { value: 'yoga', label: 'Yoga', category: 'Active' },
  { value: 'travel', label: 'Travel', category: 'Lifestyle' },
  { value: 'hiking', label: 'Hiking', category: 'Lifestyle' },
  { value: 'cooking', label: 'Cooking', category: 'Lifestyle' },
  { value: 'architecture', label: 'Architecture', category: 'Lifestyle' },
  { value: 'gaming', label: 'Gaming', category: 'Geek' },
  { value: 'anime', label: 'Anime', category: 'Geek' },
  { value: 'science', label: 'Science', category: 'Geek' },
  { value: 'technology', label: 'Technology', category: 'Geek' },
  { value: 'programming', label: 'Programming', category: 'Geek' },
  { value: 'history', label: 'History', category: 'Ideas' },
  { value: 'philosophy', label: 'Philosophy', category: 'Ideas' }
];

function renderInterestPicker(options) {
  const root = document.getElementById(options.rootId);
  if (!root) return;

  const inputName = options.inputName;
  const selected = new Set((options.selectedValues || []).map(v => String(v).toLowerCase()));

  root.innerHTML = `
    <div class="interest-picker" data-input-name="${inputName}">
      <div class="interest-picker-head">
        <input class="interest-search" type="text" placeholder="Search interests" data-role="search" />
        <div class="interest-selected" data-role="selected"></div>
      </div>
      <div class="interest-groups" data-role="groups"></div>
    </div>
  `;

  const picker = root.querySelector('.interest-picker');
  const searchEl = picker.querySelector('[data-role="search"]');
  const selectedEl = picker.querySelector('[data-role="selected"]');
  const groupsEl = picker.querySelector('[data-role="groups"]');
  const categories = Array.from(new Set(INTEREST_CATALOG.map(item => item.category)));

  function renderSelected() {
    selectedEl.innerHTML = '';
    if (selected.size === 0) {
      selectedEl.innerHTML = '<span class="interest-empty">No interests selected</span>';
      return;
    }

    for (const item of INTEREST_CATALOG.filter(it => selected.has(it.value))) {
      const chip = document.createElement('button');
      chip.type = 'button';
      chip.className = 'interest-selected-chip';
      chip.textContent = item.label;
      chip.onclick = () => {
        selected.delete(item.value);
        renderSelected();
        renderGroups(searchEl.value);
      };
      selectedEl.appendChild(chip);
    }
  }

  function renderGroups(query = '') {
    const q = String(query || '').trim().toLowerCase();
    groupsEl.innerHTML = '';

    for (const category of categories) {
      const items = INTEREST_CATALOG.filter(item =>
        item.category === category &&
        (!q || item.label.toLowerCase().includes(q) || item.value.includes(q))
      );
      if (items.length === 0) continue;

      const section = document.createElement('section');
      section.className = 'interest-group';

      const title = document.createElement('div');
      title.className = 'interest-group-title';
      title.textContent = category;
      section.appendChild(title);

      const grid = document.createElement('div');
      grid.className = 'interest-grid';

      for (const item of items) {
        const label = document.createElement('label');
        label.className = 'interest-card' + (selected.has(item.value) ? ' selected' : '');

        const input = document.createElement('input');
        input.type = 'checkbox';
        input.name = inputName;
        input.value = item.value;
        input.checked = selected.has(item.value);
        input.onchange = () => {
          if (input.checked) {
            selected.add(item.value);
          } else {
            selected.delete(item.value);
          }
          label.classList.toggle('selected', input.checked);
          renderSelected();
        };

        const text = document.createElement('span');
        text.className = 'interest-card-label';
        text.textContent = item.label;

        const meta = document.createElement('span');
        meta.className = 'interest-card-meta';
        meta.textContent = item.category;

        label.appendChild(input);
        label.appendChild(text);
        label.appendChild(meta);
        grid.appendChild(label);
      }

      section.appendChild(grid);
      groupsEl.appendChild(section);
    }
  }

  searchEl.addEventListener('input', () => renderGroups(searchEl.value));
  renderSelected();
  renderGroups();
}

window.AppInterests = {
  renderInterestPicker,
  catalog: INTEREST_CATALOG
};
