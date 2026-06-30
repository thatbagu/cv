const instance = 'https://mastodon.social';
const username = '112950868045593874';

async function fetchLatestPost() {
    const container = document.getElementById('mastodon-post');
    if (!container) return;

    try {
        const response = await fetch(`${instance}/api/v1/accounts/${username}/statuses?limit=1`);
        const data = await response.json();

        if (data.length > 0) {
            displayPost(data[0]);
        } else {
            container.innerHTML = '<p>No recent posts found.</p>';
        }
    } catch (error) {
        console.error('Error fetching Mastodon post:', error);
        container.innerHTML = '<p>No recent posts found.</p>';
    }
}

function displayPost(post) {
    const postElement = document.getElementById('mastodon-post');
    if (!postElement) return;
    postElement.innerHTML = `
        <h2>My Thoughts Today</h2>
        <div>${post.content}</div>
        <p><small>Posted on: ${new Date(post.created_at).toLocaleString()}</small></p>
    `;
}
