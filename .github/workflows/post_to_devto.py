#!/usr/bin/env python3
"""
Posts blog content to Dev.to.
Can either cross-post full articles or just share links.
"""

import os
import sys
import json
import requests
from typing import Dict, List


def read_markdown_content(file_path: str) -> str:
    """
    Read the markdown content from a blog post file (excluding frontmatter).

    Args:
        file_path: Path to markdown file

    Returns:
        Markdown content without frontmatter
    """
    try:
        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read()

        # Split by --- to remove frontmatter
        parts = content.split("---", 2)
        if len(parts) >= 3:
            return parts[2].strip()
        return content
    except Exception as e:
        print(f"✗ Error reading markdown file: {e}", file=sys.stderr)
        return ""


def format_devto_article(post_info: Dict, file_path: str, mode: str = "link") -> Dict:
    """
    Format post information into a Dev.to article.

    Args:
        post_info: Post information dict
        file_path: Path to the markdown file
        mode: "full" for full article, "link" for link post

    Returns:
        Article payload dict
    """
    title = post_info.get("title", "")
    abstract = post_info.get("abstract", "")
    url = post_info.get("url", "")

    # Get metadata
    metadata = post_info.get("metadata", {})
    tags = metadata.get("tags_devto", metadata.get("tags", []))

    # Dev.to allows max 4 tags
    tags = tags[:4]

    if mode == "full":
        # Cross-post full article
        markdown_content = read_markdown_content(file_path)

        # Create article with canonical URL pointing to your site
        article = {
            "article": {
                "title": title,
                "published": True,
                "body_markdown": markdown_content,
                "canonical_url": url,  # Important: points back to your site for SEO
                "tags": tags,
                "description": abstract,
            }
        }
    else:
        # Just share a link with description
        body = f"{abstract}\n\n[Read the full article on my blog]({url})"

        article = {
            "article": {
                "title": title,
                "published": True,
                "body_markdown": body,
                "tags": tags,
                "description": abstract,
            }
        }

    return article


def post_to_devto(api_key: str, article: Dict) -> bool:
    """
    Post an article to Dev.to using their API.

    Args:
        api_key: Dev.to API key
        article: Article payload

    Returns:
        True if successful, False otherwise
    """
    url = "https://dev.to/api/articles"

    headers = {"api-key": api_key, "Content-Type": "application/json"}

    try:
        print(f"Posting to Dev.to...", file=sys.stderr)
        print(f"Title: {article['article']['title']}", file=sys.stderr)
        print(f"Tags: {article['article']['tags']}", file=sys.stderr)

        response = requests.post(url, json=article, headers=headers, timeout=30)

        # Parse response
        try:
            result = response.json()
        except:
            result = {"text": response.text}

        if response.status_code == 201:
            article_url = result.get("url", "")
            article_id = result.get("id", "unknown")
            print(f"✓ Successfully posted to Dev.to!", file=sys.stderr)
            print(f"  Article ID: {article_id}", file=sys.stderr)
            print(f"  URL: {article_url}", file=sys.stderr)
            return True
        else:
            # Show detailed error from Dev.to
            error_msg = result.get("error", result.get("message", "Unknown error"))
            print(
                f"✗ Dev.to API error (code {response.status_code}): {error_msg}",
                file=sys.stderr,
            )
            print(f"  Full response: {result}", file=sys.stderr)

            # Common error messages
            if response.status_code == 401:
                print("  → API key is invalid", file=sys.stderr)
            elif response.status_code == 422:
                print("  → Validation error. Check article format", file=sys.stderr)

            return False

    except requests.exceptions.Timeout:
        print(f"✗ Dev.to API timeout", file=sys.stderr)
        return False
    except requests.exceptions.RequestException as e:
        print(f"✗ Failed to post to Dev.to: {e}", file=sys.stderr)
        return False


def main():
    """
    Main entry point. Reads posts from stdin (JSON) and posts to Dev.to.
    """
    import argparse

    parser = argparse.ArgumentParser(description="Post blog content to Dev.to")
    parser.add_argument("--api-key", required=True, help="Dev.to API key")
    parser.add_argument(
        "--mode",
        choices=["full", "link"],
        default="link",
        help='Posting mode: "full" for full article cross-post, "link" for link with excerpt',
    )
    parser.add_argument(
        "--posts-json",
        help="JSON file with posts to publish (default: read from stdin)",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Print article data without actually posting",
    )

    args = parser.parse_args()

    # Read posts data
    if args.posts_json:
        with open(args.posts_json, "r") as f:
            posts = json.load(f)
    else:
        posts = json.load(sys.stdin)

    if not posts:
        print("No posts to publish", file=sys.stderr)
        return 0

    print(
        f"Mode: {args.mode} - {'Cross-posting full articles' if args.mode == 'full' else 'Sharing links with excerpts'}",
        file=sys.stderr,
    )

    # Track successful posts
    successful_posts = []

    # Post each article that needs Dev.to posting
    for file_path, post_info in posts.items():
        platforms = post_info.get("platforms", [])

        if "devto" not in platforms:
            print(
                f"○ Skipping {file_path} (Dev.to not in platforms list)",
                file=sys.stderr,
            )
            continue

        print(f"\n{'=' * 60}", file=sys.stderr)
        print(f"Processing: {file_path}", file=sys.stderr)
        print(f"Title: {post_info.get('title')}", file=sys.stderr)
        print(f"URL: {post_info.get('url')}", file=sys.stderr)

        article = format_devto_article(post_info, file_path, args.mode)

        if args.dry_run:
            print("\n--- DRY RUN: Would post this article ---", file=sys.stderr)
            print(json.dumps(article, indent=2), file=sys.stderr)
            print("--- END ARTICLE ---\n", file=sys.stderr)
            successful_posts.append(file_path)
        else:
            success = post_to_devto(args.api_key, article)

            if success:
                successful_posts.append(file_path)
            else:
                print(f"✗ Failed to post {file_path}", file=sys.stderr)

    # Output successful posts as JSON for the next step
    print(
        f"\n=== Posted {len(successful_posts)} article(s) to Dev.to ===",
        file=sys.stderr,
    )

    # Return JSON array of successfully posted files
    output = {"successful_posts": successful_posts, "platform": "devto"}

    # Write to file for GitHub Actions to pick up
    with open("devto_results.json", "w") as f:
        json.dump(output, f, indent=2)

    # Always return 0 - no posts isn't an error, just means nothing to do
    return 0


if __name__ == "__main__":
    sys.exit(main())
