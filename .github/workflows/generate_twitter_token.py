#!/usr/bin/env python3
"""
One-time local script to generate a Twitter OAuth 2.0 refresh token.
Run this once, store the refresh token in GitHub secrets as TWITTER_REFRESH_TOKEN.
CI will use it to get fresh access tokens automatically.

Usage:
    pip install tweepy
    python .github/workflows/generate_twitter_token.py \
        --client-id YOUR_CLIENT_ID \
        --client-secret YOUR_CLIENT_SECRET
"""

import argparse
import tweepy

parser = argparse.ArgumentParser()
parser.add_argument("--client-id", required=True)
parser.add_argument("--client-secret", required=True)
args = parser.parse_args()

handler = tweepy.OAuth2UserHandler(
    client_id=args.client_id,
    client_secret=args.client_secret,
    redirect_uri="https://localhost",
    scope=["tweet.read", "tweet.write", "users.read", "offline.access"],
)

print("\n1. Open this URL in your browser and authorize the app:")
print(handler.get_authorization_url())
print("\n2. After authorizing, paste the full redirect URL below.\n")

redirect_url = input("Redirect URL: ").strip()
token = handler.fetch_token(redirect_url)

print("\n✓ Done! Set this as TWITTER_REFRESH_TOKEN in GitHub secrets:\n")
print(token["refresh_token"])
print()
