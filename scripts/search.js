document.addEventListener('DOMContentLoaded', () => {
  const searchInput = document.getElementById('search-input');
  const searchBtn = document.getElementById('search-btn');
  const searchResults = document.getElementById('search-results');

  let pendingHighlight = null;

  function performSearch() {
    const term = searchInput.value.trim();
    if (!term) return;

    searchResults.innerHTML = '<div class="search-status">Searching...</div>';

    fetch(`/search?term=${encodeURIComponent(term)}&page=all`)
      .then(r => r.json())
      .then(data => {
        if (data.results.length === 0) {
          searchResults.innerHTML = `<div class="search-status">No results for "${term}"</div>`;
          return;
        }

        const items = data.results.map((result, i) => {
          const label = result.file.endsWith('.md') ? result.title : result.file.replace('pages/', '').replace('.html', '');
          const excerpt = result.excerpt ? ` — ${result.excerpt.slice(0, 80)}...` : '';
          return `<div class="search-result-item" data-path="${result.path}">
            <span class="search-result-label">${i + 1}. ${label}</span><span>${excerpt}</span>
          </div>`;
        }).join('');

        searchResults.innerHTML = items;

        searchResults.querySelectorAll('.search-result-item').forEach(item => {
          item.addEventListener('click', () => {
            const path = item.dataset.path;
            pendingHighlight = { searchTerm: data.searchTerm };
            htmx.ajax('GET', path, { target: '#main-content', swap: 'innerHTML' });
            history.pushState(null, '', path);
            searchResults.innerHTML = '';
            searchInput.value = '';
          });
        });
      })
      .catch(() => {
        searchResults.innerHTML = '<div class="search-status">Search failed. Please try again.</div>';
      });
  }

  searchInput.addEventListener('keydown', e => {
    if (e.key === 'Enter') performSearch();
    if (e.key === 'Escape') searchResults.innerHTML = '';
  });

  searchBtn.addEventListener('click', performSearch);

  document.addEventListener('click', e => {
    if (!e.target.closest('#search-bar') && !e.target.closest('#search-results')) {
      searchResults.innerHTML = '';
    }
  });

  document.body.addEventListener('htmx:afterSwap', () => {
    if (pendingHighlight) {
      highlightSearchTerms(pendingHighlight.searchTerm);
      pendingHighlight = null;
    }
  });

  function highlightSearchTerms(searchTerm) {
    const mainContent = document.getElementById('main-content');
    if (!mainContent) return;

    const regex = new RegExp(escapeRegExp(searchTerm), 'gi');
    const walker = document.createTreeWalker(mainContent, NodeFilter.SHOW_TEXT);

    let node;
    while ((node = walker.nextNode())) {
      if (regex.test(node.nodeValue)) {
        const span = document.createElement('span');
        span.innerHTML = node.nodeValue.replace(regex, '<mark>$&</mark>');
        node.parentNode.replaceChild(span, node);
      }
    }

    const firstMark = mainContent.querySelector('mark');
    if (firstMark) firstMark.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }

  function escapeRegExp(string) {
    return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  }
});
