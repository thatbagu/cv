#!/usr/bin/env python3
"""
Posts blog content to Mastodon.
Reads from detect_posts_to_publish.py output and posts to Mastodon.
"""

import os
import sys
import json
import requests
from typing import Dict, List


def format_mastodon_message(post_info: Dict) -> str:
    """
    Format post information into a Mastodon post.
    Mastodon allows 500 characters by default (can be more on some instances).

    Args:
        post_info: Post information dict

    Returns:
        Formatted Mastodon post text
    """
    title = post_info.get("title", "")
    abstract = post_info.get("abstract", "")
    url = post_info.get("url", "")

    # Get hashtags from metadata, or use defaults
    metadata = post_info.get("metadata", {})
    hashtags = metadata.get("hashtags", ["TechBlog", "Programming", "WebDev"])

    # Format hashtags with # prefix
    hashtag_string = " ".join(
        [f"#{tag}" if not tag.startswith("#") else tag for tag in hashtags]
    )

    # Mastodon has 500 char limit on most instances
    # Build the post
    post = f"📝 {title}\n\n{abstract}\n\n{url}\n\n{hashtag_string}"

    # Check if we're over limit
    if len(post) > 500:
        # Try without hashtags first
        post_no_tags = f"📝 {title}\n\n{abstract}\n\n{url}"
        if len(post_no_tags) <= 500:
            post = post_no_tags
        else:
            # Truncate abstract to fit
            available = 500 - len(f"📝 {title}\n\n\n\n{url}") - 3
            truncated_abstract = abstract[:available] + "..."
            post = f"📝 {title}\n\n{truncated_abstract}\n\n{url}"

    return post


def post_to_mastodon(instance_url: str, access_token: str, message: str) -> bool:
    """
    Post a status to Mastodon using the API.

    Args:
        instance_url: Mastodon instance URL (e.g., https://mastodon.social)
        access_token: Mastodon access token
        message: Status text to post

    Returns:
        True if successful, False otherwise
    """
    # Remove trailing slash from instance URL if present
    instance_url = instance_url.rstrip("/")

    url = f"{instance_url}/api/v1/statuses"

    headers = {
        "Authorization": f"Bearer {access_token}",
        "Content-Type": "application/json",
    }

    payload = {
        "status": message,
        "visibility": "public",  # Options: public, unlisted, private, direct
    }

    try:
        print(f"Posting to Mastodon ({instance_url})...", file=sys.stderr)
        print(f"Post length: {len(message)} characters", file=sys.stderr)

        response = requests.post(url, json=payload, headers=headers, timeout=10)

        # Parse response
        try:
            result = response.json()
        except:
            result = {"text": response.text}

        if response.status_code == 200:
            post_id = result.get("id", "unknown")
            post_url = result.get("url", "")
            print(f"✓ Successfully posted to Mastodon!", file=sys.stderr)
            print(f"  Post ID: {post_id}", file=sys.stderr)
            if post_url:
                print(f"  URL: {post_url}", file=sys.stderr)
            return True
        else:
            # Show detailed error from Mastodon
            error_msg = result.get("error", "Unknown error")
            print(
                f"✗ Mastodon API error (code {response.status_code}): {error_msg}",
                file=sys.stderr,
            )
            print(f"  Full response: {result}", file=sys.stderr)

            # Common error messages
            if response.status_code == 401:
                print("  → Access token is invalid or expired", file=sys.stderr)
            elif response.status_code == 403:
                print("  → Permission denied. Check token permissions", file=sys.stderr)
            elif response.status_code == 422:
                print("  → Validation error. Check message content", file=sys.stderr)

            return False

    except requests.exceptions.Timeout:
        print(f"✗ Mastodon API timeout", file=sys.stderr)
        return False
    except requests.exceptions.RequestException as e:
        print(f"✗ Failed to post to Mastodon: {e}", file=sys.stderr)
        return False


def main():
    """
    Main entry point. Reads posts from stdin (JSON) and posts to Mastodon.
    """
    import argparse

    parser = argparse.ArgumentParser(description="Post blog content to Mastodon")
    parser.add_argument(
        "--instance-url",
        required=True,
        help="Mastodon instance URL (e.g., https://mastodon.social)",
    )
    parser.add_argument("--access-token", required=True, help="Mastodon access token")
    parser.add_argument(
        "--posts-json",
        help="JSON file with posts to publish (default: read from stdin)",
    )
    parser.add_argument(
        "--dry-run", action="store_true", help="Print posts without actually posting"
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

    print(f"Found {len(posts)} post(s) in JSON", file=sys.stderr)

    # Track successful posts
    successful_posts = []

    # Post each article that needs Mastodon posting
    for file_path, post_info in posts.items():
        print(f"\nChecking {file_path}...", file=sys.stderr)
        platforms = post_info.get("platforms", [])
        print(f"  Platforms for this post: {platforms}", file=sys.stderr)

        if "mastodon" not in platforms:
            print(
                f"○ Skipping {file_path} (Mastodon not in platforms list)",
                file=sys.stderr,
            )
            continue

        print(f"\n{'=' * 60}", file=sys.stderr)
        print(f"Processing: {file_path}", file=sys.stderr)
        print(f"Title: {post_info.get('title')}", file=sys.stderr)
        print(f"URL: {post_info.get('url')}", file=sys.stderr)

        message = format_mastodon_message(post_info)

        if args.dry_run:
            print("\n--- DRY RUN: Would post this to Mastodon ---", file=sys.stderr)
            print(message, file=sys.stderr)
            print(f"--- END POST ({len(message)} chars) ---\n", file=sys.stderr)
            successful_posts.append(file_path)
        else:
            print("\n--- Post content ---", file=sys.stderr)
            print(message, file=sys.stderr)
            print("--- End content ---\n", file=sys.stderr)

            success = post_to_mastodon(args.instance_url, args.access_token, message)

            if success:
                successful_posts.append(file_path)
            else:
                print(f"✗ Failed to post {file_path}", file=sys.stderr)

    # Output successful posts as JSON for the next step
    print(
        f"\n=== Posted {len(successful_posts)} article(s) to Mastodon ===",
        file=sys.stderr,
    )

    # Return JSON array of successfully posted files
    output = {"successful_posts": successful_posts, "platform": "mastodon"}

    # Write to file for GitHub Actions to pick up
    with open("mastodon_results.json", "w") as f:
        json.dump(output, f, indent=2)

    # Always return 0 - no posts isn't an error, just means nothing to do
    return 0


if __name__ == "__main__":
    sys.exit(main())
