#!/usr/bin/env python3
"""
Posts blog content to Telegram channel.
Reads from detect_posts_to_publish.py output and posts to Telegram.
"""

import os
import sys
import json
import requests
from typing import Dict, List


def format_telegram_message(post_info: Dict) -> str:
    title = post_info.get("title", "")
    abstract = post_info.get("abstract", "")
    excerpt = post_info.get("excerpt", "")
    url = post_info.get("url", "")

    metadata = post_info.get("metadata", {})
    hashtags = metadata.get("hashtags_telegram", metadata.get("hashtags", []))
    hashtag_string = " ".join(
        [f"#{tag}" if not tag.startswith("#") else tag for tag in hashtags]
    )

    message = f"<b>{title}</b>\n\n{abstract}"

    if excerpt:
        message += f"\n\n{excerpt}"

    message += f"\n\n<a href=\"{url}\">Read full article →</a>"

    if hashtag_string:
        message += f"\n\n{hashtag_string}"

    return message


def post_to_telegram(bot_token: str, channel_id: str, message: str) -> bool:
    """
    Post a message to Telegram channel using Bot API.

    Args:
        bot_token: Telegram bot token from @BotFather
        channel_id: Channel ID (e.g., @your_channel or -1001234567890)
        message: Message text to send

    Returns:
        True if successful, False otherwise
    """
    url = f"https://api.telegram.org/bot{bot_token}/sendMessage"

    payload = {
        "chat_id": channel_id,
        "text": message,
        "parse_mode": "HTML",
        "disable_web_page_preview": False,  # Show preview for the link
    }

    try:
        print(f"Posting to Telegram channel: {channel_id}", file=sys.stderr)
        print(f"Message length: {len(message)} characters", file=sys.stderr)

        response = requests.post(url, json=payload, timeout=10)

        # Try to parse response even if status code is error
        try:
            result = response.json()
        except:
            result = {}

        if response.status_code == 200 and result.get("ok"):
            print(f"✓ Successfully posted to Telegram", file=sys.stderr)
            return True
        else:
            # Show detailed error from Telegram
            error_desc = result.get("description", "No error description")
            error_code = result.get("error_code", response.status_code)
            print(
                f"✗ Telegram API error (code {error_code}): {error_desc}",
                file=sys.stderr,
            )
            print(f"  Channel ID used: {channel_id}", file=sys.stderr)
            print(f"  Response: {result}", file=sys.stderr)
            return False

    except requests.exceptions.Timeout:
        print(f"✗ Telegram API timeout", file=sys.stderr)
        return False
    except requests.exceptions.RequestException as e:
        print(f"✗ Failed to post to Telegram: {e}", file=sys.stderr)
        return False


def main():
    """
    Main entry point. Reads posts from stdin (JSON) and posts to Telegram.
    """
    import argparse

    parser = argparse.ArgumentParser(
        description="Post blog content to Telegram channel"
    )
    parser.add_argument("--bot-token", required=True, help="Telegram bot token")
    parser.add_argument(
        "--channel-id",
        required=True,
        help="Telegram channel ID (e.g., @channel_name or -1001234567890)",
    )
    parser.add_argument(
        "--posts-json",
        help="JSON file with posts to publish (default: read from stdin)",
    )
    parser.add_argument(
        "--dry-run", action="store_true", help="Print messages without actually posting"
    )

    args = parser.parse_args()

    # Read posts data
    if args.posts_json:
        with open(args.posts_json, "r") as f:
            posts = json.load(f)
    else:
        posts = json.load(sys.stdin)

    if not posts:
        print("No posts to publish")
        return True

    # Track successful posts
    successful_posts = []

    # Post each article that needs Telegram posting
    for file_path, post_info in posts.items():
        platforms = post_info.get("platforms", [])

        if "telegram" not in platforms:
            print(
                f"○ Skipping {file_path} (Telegram not in platforms list)",
                file=sys.stderr,
            )
            continue

        print(f"\n{'=' * 60}", file=sys.stderr)
        print(f"Processing: {file_path}", file=sys.stderr)
        print(f"Title: {post_info.get('title')}", file=sys.stderr)
        print(f"URL: {post_info.get('url')}", file=sys.stderr)

        message = format_telegram_message(post_info)

        if args.dry_run:
            print("\n--- DRY RUN: Would post this message ---", file=sys.stderr)
            print(message, file=sys.stderr)
            print("--- END MESSAGE ---\n", file=sys.stderr)
            successful_posts.append(file_path)
        else:
            print("\n--- Message to post ---", file=sys.stderr)
            print(message, file=sys.stderr)
            print("--- End message ---\n", file=sys.stderr)

            success = post_to_telegram(args.bot_token, args.channel_id, message)

            if success:
                successful_posts.append(file_path)
            else:
                print(f"✗ Failed to post {file_path}", file=sys.stderr)

    # Output successful posts as JSON for the next step
    print(f"\n=== Posted {len(successful_posts)} article(s) to Telegram ===")

    # Return JSON array of successfully posted files
    output = {"successful_posts": successful_posts, "platform": "telegram"}

    # Write to file for GitHub Actions to pick up
    with open("telegram_results.json", "w") as f:
        json.dump(output, f, indent=2)

    return 0


if __name__ == "__main__":
    sys.exit(main())
