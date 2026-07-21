const instance = 'https://mastodon.social';
const username = '112950868045593874';

async function fetchLatestPost() {
    const container = document.getElementById('mastodon-post');
    if (!container) return;

    try {
        const response = await fetch(`${instance}/api/v1/accounts/${username}/statuses?limit=3&exclude_replies=true&exclude_reblogs=true`);
        const data = await response.json();

        if (data.length > 0) {
            displayPosts(data);
        } else {
            container.innerHTML = '<p>No recent posts found.</p>';
        }
    } catch (error) {
        console.error('Error fetching Mastodon posts:', error);
        container.innerHTML = '<p>No recent posts found.</p>';
    }
}

function displayPosts(posts) {
    const container = document.getElementById('mastodon-post');
    if (!container) return;

    const items = posts.map(post => `
        <div class="mastodon-post-item">
            <div class="mastodon-post-content">${post.content}</div>
            <p class="mastodon-post-date"><small>${new Date(post.created_at).toLocaleString()}</small></p>
        </div>
    `).join('');

    container.innerHTML = `<h2>Recent posts</h2>${items}`;
}
