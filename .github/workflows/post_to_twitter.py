#!/usr/bin/env python3
"""
Posts blog content to Twitter/X.
Reads from detect_posts_to_publish.py output and posts tweets.
"""

import os
import sys
import json
import requests
from typing import Dict, List


def format_twitter_message(post_info: Dict, max_length: int = 280) -> str:
    """
    Format post information into a tweet.
    Twitter has 280 character limit. Links count as ~23 characters.

    Args:
        post_info: Post information dict
        max_length: Maximum tweet length (default 280)

    Returns:
        Formatted tweet text
    """
    title = post_info.get("title", "")
    abstract = post_info.get("abstract", "")
    url = post_info.get("url", "")

    # Get hashtags from metadata - try Twitter-specific first, then general
    metadata = post_info.get("metadata", {})
    hashtags = metadata.get("hashtags_twitter", metadata.get("hashtags", []))

    # Format hashtags
    hashtag_string = " ".join(
        [f"#{tag}" if not tag.startswith("#") else tag for tag in hashtags]
    )

    # Calculate available space
    # URL ~23 chars, emoji 2, newlines, hashtags
    reserved = 23 + 2 + 4  # URL + emoji + newlines
    if hashtag_string:
        reserved += len(hashtag_string) + 2  # hashtags + newlines

    available_chars = max_length - reserved

    # Build tweet
    tweet = f"📝 {title}\n\n"

    # Add abstract if it fits
    remaining = available_chars - len(tweet)
    if len(abstract) <= remaining:
        tweet += f"{abstract}\n\n"
    else:
        # Truncate abstract
        truncated = abstract[: remaining - 4] + "...\n\n"
        tweet += truncated

    tweet += url

    # Add hashtags if they fit
    if hashtag_string and len(tweet) + len(hashtag_string) + 2 <= max_length:
        tweet += f"\n\n{hashtag_string}"

    return tweet


def post_to_twitter_v2(
    bearer_token: str,
    access_token: str,
    access_token_secret: str,
    api_key: str,
    api_secret: str,
    message: str,
) -> bool:
    """
    Post a tweet using Twitter API v2 with OAuth 1.0a User Context.

    Args:
        bearer_token: App-only bearer token (not used for posting, but kept for compatibility)
        access_token: OAuth 1.0a Access Token
        access_token_secret: OAuth 1.0a Access Token Secret
        api_key: OAuth 1.0a API Key (Consumer Key)
        api_secret: OAuth 1.0a API Secret (Consumer Secret)
        message: Tweet text to post

    Returns:
        True if successful, False otherwise
    """
    from requests_oauthlib import OAuth1

    url = "https://api.twitter.com/2/tweets"

    # OAuth 1.0a authentication
    auth = OAuth1(api_key, api_secret, access_token, access_token_secret)

    payload = {"text": message}

    try:
        print(f"Posting to Twitter/X...", file=sys.stderr)
        print(f"Tweet length: {len(message)} characters", file=sys.stderr)
        print(f"Tweet preview:\n{message}\n", file=sys.stderr)

        response = requests.post(url, json=payload, auth=auth, timeout=10)

        # Parse response
        try:
            result = response.json()
        except:
            result = {}

        if response.status_code == 201:  # Twitter returns 201 for successful creation
            tweet_id = result.get("data", {}).get("id", "unknown")
            print(
                f"✓ Successfully posted to Twitter! Tweet ID: {tweet_id}",
                file=sys.stderr,
            )
            return True
        else:
            # Show detailed error from Twitter
            error_title = result.get("title", "Unknown error")
            error_detail = result.get("detail", "No error details")
            errors = result.get("errors", [])

            print(
                f"✗ Twitter API error (code {response.status_code}): {error_title}",
                file=sys.stderr,
            )
            print(f"  Detail: {error_detail}", file=sys.stderr)
            if errors:
                for err in errors:
                    print(f"  - {err.get('message', err)}", file=sys.stderr)
            print(f"  Full response: {result}", file=sys.stderr)
            return False

    except ImportError:
        print(
            f"✗ requests-oauthlib not installed. Install with: pip install requests-oauthlib",
            file=sys.stderr,
        )
        return False
    except requests.exceptions.Timeout:
        print(f"✗ Twitter API timeout", file=sys.stderr)
        return False
    except requests.exceptions.RequestException as e:
        print(f"✗ Failed to post to Twitter: {e}", file=sys.stderr)
        return False


def main():
    """
    Main entry point. Reads posts from stdin (JSON) and posts to Twitter.
    """
    import argparse

    parser = argparse.ArgumentParser(description="Post blog content to Twitter/X")
    parser.add_argument(
        "--api-key", required=True, help="Twitter API Key (Consumer Key)"
    )
    parser.add_argument(
        "--api-secret", required=True, help="Twitter API Secret (Consumer Secret)"
    )
    parser.add_argument("--access-token", required=True, help="Twitter Access Token")
    parser.add_argument(
        "--access-token-secret", required=True, help="Twitter Access Token Secret"
    )
    parser.add_argument(
        "--bearer-token", help="Twitter Bearer Token (optional, not used for posting)"
    )
    parser.add_argument(
        "--posts-json",
        help="JSON file with posts to publish (default: read from stdin)",
    )
    parser.add_argument(
        "--dry-run", action="store_true", help="Print tweets without actually posting"
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

    # Track successful posts
    successful_posts = []

    # Post each article that needs Twitter posting
    for file_path, post_info in posts.items():
        platforms = post_info.get("platforms", [])

        if "twitter" not in platforms:
            print(
                f"○ Skipping {file_path} (Twitter not in platforms list)",
                file=sys.stderr,
            )
            continue

        print(f"\n{'=' * 60}", file=sys.stderr)
        print(f"Processing: {file_path}", file=sys.stderr)
        print(f"Title: {post_info.get('title')}", file=sys.stderr)
        print(f"URL: {post_info.get('url')}", file=sys.stderr)

        message = format_twitter_message(post_info)

        if args.dry_run:
            print("\n--- DRY RUN: Would post this tweet ---", file=sys.stderr)
            print(message, file=sys.stderr)
            print(f"--- END TWEET ({len(message)} chars) ---\n", file=sys.stderr)
            successful_posts.append(file_path)
        else:
            success = post_to_twitter_v2(
                args.bearer_token or "",
                args.access_token,
                args.access_token_secret,
                args.api_key,
                args.api_secret,
                message,
            )

            if success:
                successful_posts.append(file_path)
            else:
                print(f"✗ Failed to post {file_path}", file=sys.stderr)

    # Output successful posts as JSON for the next step
    print(
        f"\n=== Posted {len(successful_posts)} article(s) to Twitter ===",
        file=sys.stderr,
    )

    # Return JSON array of successfully posted files
    output = {"successful_posts": successful_posts, "platform": "twitter"}

    # Write to file for GitHub Actions to pick up
    with open("twitter_results.json", "w") as f:
        json.dump(output, f, indent=2)

    # Always return 0 - no posts isn't an error, just means nothing to do
    return 0


if __name__ == "__main__":
    sys.exit(0 if main() else 1)
