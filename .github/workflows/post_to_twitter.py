#!/usr/bin/env python3

import sys
import json
import tweepy
from typing import Dict


def get_access_token(client_id: str, client_secret: str, refresh_token: str) -> str:
    handler = tweepy.OAuth2UserHandler(
        client_id=client_id,
        client_secret=client_secret,
        redirect_uri="https://localhost",
        scope=["tweet.read", "tweet.write", "users.read", "offline.access"],
    )
    token = handler.refresh_token(
        "https://api.twitter.com/2/oauth2/token",
        refresh_token=refresh_token,
    )
    return token["access_token"]


def format_tweet(post_info: Dict, max_length: int = 280) -> str:
    title = post_info.get("title", "")
    abstract = post_info.get("abstract", "")
    url = post_info.get("url", "")

    metadata = post_info.get("metadata", {})
    hashtags = metadata.get("hashtags_twitter", metadata.get("hashtags", []))
    hashtag_string = " ".join(
        f"#{tag}" if not tag.startswith("#") else tag for tag in hashtags
    )

    suffix = f"\n\n{url}"
    if hashtag_string:
        suffix += f"\n\n{hashtag_string}"

    available = max_length - 23 - len(suffix) + len(url)
    header = f"{title}\n\n"
    body = (
        abstract
        if len(header) + len(abstract) <= available
        else abstract[: available - len(header) - 3] + "..."
    )

    return header + body + suffix


def post_tweet(client: tweepy.Client, text: str) -> bool:
    print(f"Tweet preview ({len(text)} chars):\n{text}\n", file=sys.stderr)
    try:
        response = client.create_tweet(text=text)
        tweet_id = response.data["id"]
        print(f"✓ Posted: https://x.com/i/web/status/{tweet_id}", file=sys.stderr)
        return True
    except tweepy.errors.TweepyException as e:
        print(f"✗ Twitter error: {e}", file=sys.stderr)
        return False


def main():
    import argparse

    parser = argparse.ArgumentParser()
    parser.add_argument("--client-id", required=True)
    parser.add_argument("--client-secret", required=True)
    parser.add_argument("--refresh-token", required=True)
    parser.add_argument("--posts-json")
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()

    posts = json.load(open(args.posts_json)) if args.posts_json else json.load(sys.stdin)
    if not posts:
        print("No posts to publish", file=sys.stderr)
        return 0

    print("Refreshing Twitter access token...", file=sys.stderr)
    access_token = get_access_token(args.client_id, args.client_secret, args.refresh_token)
    client = tweepy.Client(access_token=access_token)

    successful_posts = []

    for file_path, post_info in posts.items():
        if "twitter" not in post_info.get("platforms", []):
            print(f"○ Skipping {file_path}", file=sys.stderr)
            continue

        print(f"\n{'=' * 60}", file=sys.stderr)
        print(f"Processing: {file_path}", file=sys.stderr)

        tweet = format_tweet(post_info)

        if args.dry_run:
            print(f"DRY RUN:\n{tweet}", file=sys.stderr)
            successful_posts.append(file_path)
        elif post_tweet(client, tweet):
            successful_posts.append(file_path)
        else:
            print(f"✗ Failed: {file_path}", file=sys.stderr)

    print(f"\n=== Posted {len(successful_posts)} article(s) to Twitter ===", file=sys.stderr)

    with open("twitter_results.json", "w") as f:
        json.dump({"successful_posts": successful_posts, "platform": "twitter"}, f, indent=2)

    return 0


if __name__ == "__main__":
    sys.exit(main())
