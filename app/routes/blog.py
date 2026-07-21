from fastapi import APIRouter, Request, HTTPException
from fastapi.responses import HTMLResponse, FileResponse
from bs4 import BeautifulSoup
import asyncio
import glob
from app.utils import read_file, parse_markdown_file, generate_blog_html

router = APIRouter()


async def get_blog_posts():
    blog_posts = []
    blog_post_files = glob.glob("blog_posts/*.md")
    for file_path in blog_post_files:
        try:
            post = await asyncio.to_thread(parse_markdown_file, file_path)
            blog_posts.append(post)
        except Exception as e:
            print(f"Error parsing blog post {file_path}: {str(e)}")
    return sorted(blog_posts, key=lambda x: x["date"], reverse=True)


@router.get("/blog/{slug}", response_class=HTMLResponse)
async def serve_blog_post(request: Request, slug: str):
    posts = await get_blog_posts()
    post = next((p for p in posts if p["slug"] == slug), None)
    if post:
        post_url = f"{{{{ window.location.origin }}}}/blog/{post['slug']}"
        blog_post_content = f"""
        <article class="blog-post">
            <div class="post-header">
                <h1>{post["title"]}</h1>
                <p class="post-date">{post["date"]}</p>
            </div>
            <div class="post-content">{post["content"]}</div>
            <div class="post-share">
                <span class="share-label">Share:</span>
                <a class="share-btn" id="shareTwitter" href="#" target="_blank" rel="noopener noreferrer" title="Share on X / Twitter">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M23 3a10.9 10.9 0 0 1-3.14 1.53 4.48 4.48 0 0 0-7.86 3v1A10.66 10.66 0 0 1 3 4s-4 9 5 13a11.64 11.64 0 0 1-7 2c9 5 20 0 20-11.5a4.5 4.5 0 0 0-.08-.83A7.72 7.72 0 0 0 23 3z"/></svg>
                </a>
                <a class="share-btn" id="shareLinkedIn" href="#" title="Share on LinkedIn">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M16 8a6 6 0 0 1 6 6v7h-4v-7a2 2 0 0 0-2-2 2 2 0 0 0-2 2v7h-4v-7a6 6 0 0 1 6-6z"/><rect x="2" y="9" width="4" height="12"/><circle cx="4" cy="4" r="2"/></svg>
                </a>
                <button class="share-btn" onclick="shareMastodon()" title="Share on Mastodon">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21.58 13.913c-.29 1.469-2.592 3.121-5.238 3.396-1.379.184-2.737.368-4.185.276-2.368-.092-4.237-.551-4.237-.551 0 .225.014.439.043.642.308 2.294 2.317 2.432 4.222 2.494 1.924.031 3.641-.46 3.641-.46l.079 1.708s-1.344.806-3.736.959c-1.32.086-2.96-.031-4.87-.511-4.145-1.023-4.867-5.139-4.975-9.323-.03-1.043-.012-2.025-.012-2.847 0-3.574 2.341-4.622 2.341-4.622 1.18-.54 3.204-.771 5.311-.788h.051c2.108.017 4.133.248 5.312.788 0 0 2.341 1.048 2.341 4.622 0 0 .029 2.634-.088 4.454z"/><path d="M17.832 7.674c0-1.023-.828-1.855-1.847-1.855-1.021 0-1.848.832-1.848 1.855 0 1.023.827 1.854 1.848 1.854 1.019 0 1.847-.831 1.847-1.854z"/></svg>
                </button>
                <button class="share-btn" id="copyLinkBtn" onclick="copyPostLink()" title="Copy link">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
                </button>
                <button class="share-btn" onclick="downloadPDF()" title="Download as PDF">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
                </button>
            </div>
        </article>
        <script>
        (function() {{
            var slug = '{post["slug"]}';
            var title = {repr(post["title"])};
            var postURL = window.location.origin + '/blog/' + slug;

            document.getElementById('shareTwitter').href =
                'https://twitter.com/intent/tweet?url=' + encodeURIComponent(postURL) + '&text=' + encodeURIComponent(title);
            document.getElementById('shareLinkedIn').href =
                'https://www.linkedin.com/sharing/share-offsite/?url=' + encodeURIComponent(postURL);
        }})();

        function copyPostLink() {{
            var postURL = window.location.origin + '/blog/{post["slug"]}';
            navigator.clipboard.writeText(postURL).then(function() {{
                var btn = document.getElementById('copyLinkBtn');
                var orig = btn.innerHTML;
                btn.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>';
                setTimeout(function() {{ btn.innerHTML = orig; }}, 2000);
            }}).catch(function(err) {{
                console.error('Failed to copy: ', err);
            }});
        }}

        function shareMastodon() {{
            var instance = prompt('Your Mastodon instance:', 'mastodon.social');
            if (!instance) return;
            instance = instance.trim().replace(/^https?:\\/\\//, '');
            var postURL = window.location.origin + '/blog/{post["slug"]}';
            var text = {repr(post["title"])} + ' ' + postURL;
            window.open('https://' + instance + '/share?text=' + encodeURIComponent(text), '_blank');
        }}

        function downloadPDF() {{
            window.print();
        }}
        </script>
        """

        if request.headers.get("HX-Request"):
            return HTMLResponse(content=blog_post_content)
        else:
            with open("index.html", "r") as file:
                template = file.read()
            soup = BeautifulSoup(template, "html.parser")

            # Update OG meta tags
            soup.find("meta", property="og:title")["content"] = post["title"]
            soup.find("meta", property="og:description")["content"] = post.get(
                "excerpt", ""
            )
            soup.find("meta", property="og:image")["content"] = (
                f"https://mlship.dev/assets/opengraph/images/{slug}.png"
            )
            soup.find("meta", property="og:url")["content"] = (
                f"https://mlship.dev/blog/{slug}"
            )
            soup.find("meta", property="og:type")["content"] = "article"

            main_content = soup.find(id="main-content")
            if main_content:
                main_content.clear()
                main_content.append(BeautifulSoup(blog_post_content, "html.parser"))

            return HTMLResponse(content=str(soup))
    else:
        raise HTTPException(status_code=404, detail="Not Found")


@router.get("/api/blog-posts")
async def serve_blog_posts():
    posts = await get_blog_posts()
    return [{k: v for k, v in post.items() if k != "content"} for post in posts]


@router.get("/atom.xml")
async def serve_atom_feed():
    """Serve the pregenerated atom.xml file"""
    try:
        return FileResponse(
            "atom.xml",
            media_type="application/atom+xml",
            headers={"Cache-Control": "public, max-age=3600"},  # Cache for 1 hour
        )
    except FileNotFoundError:
        raise HTTPException(status_code=404, detail="Atom feed not found")


@router.get("/rss.xml")
async def serve_rss_feed():
    """Serve the pregenerated rss.xml file"""
    try:
        return FileResponse(
            "rss.xml",
            media_type="application/rss+xml",
            headers={"Cache-Control": "public, max-age=3600"},  # Cache for 1 hour
        )
    except FileNotFoundError:
        raise HTTPException(status_code=404, detail="RSS feed not found")


@router.get("/feed")
@router.get("/feed.xml")
async def serve_default_feed():
    """Redirect common feed URLs to RSS feed"""
    try:
        return FileResponse(
            "rss.xml",
            media_type="application/rss+xml",
            headers={"Cache-Control": "public, max-age=3600"},
        )
    except FileNotFoundError:
        raise HTTPException(status_code=404, detail="Feed not found")
