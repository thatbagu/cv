#!/usr/bin/env python3
"""
Detects which blog posts need to be cross-posted to social media platforms.
Tracks posting status in markdown frontmatter via 'posted_to' field.
"""

import glob
import yaml
import subprocess
import sys
from datetime import datetime
from typing import List, Dict, Set


def get_changed_markdown_files() -> List[str]:
    """
    Get list of markdown files that changed in the last commit.
    Returns absolute paths to changed .md files in blog_posts/
    """
    try:
        # Get files changed in the last commit
        result = subprocess.run(
            ["git", "diff", "--name-only", "HEAD~1", "HEAD"],
            capture_output=True,
            text=True,
            check=True,
        )

        changed_files = result.stdout.strip().split("\n")

        # Filter out empty strings and filter for markdown files in blog_posts directory
        markdown_files = [
            f
            for f in changed_files
            if f and f.startswith("blog_posts/") and f.endswith(".md")
        ]

        return markdown_files
    except subprocess.CalledProcessError as e:
        # If git command fails, try to get all markdown files (fallback for first commit)
        # Log to stderr only
        return glob.glob("blog_posts/*.md")


def parse_frontmatter(file_path: str) -> Dict:
    """
    Parse frontmatter from markdown file.
    Returns metadata dictionary with 'posted_to' field (defaults to empty list).
    """
    try:
        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read()

        # Split by --- to extract frontmatter
        parts = content.split("---", 2)
        if len(parts) < 3:
            print(
                f"Warning: {file_path} has invalid frontmatter format", file=sys.stderr
            )
            return {}

        frontmatter = parts[1].strip()
        metadata = yaml.safe_load(frontmatter)

        # Ensure posted_to field exists
        if "posted_to" not in metadata:
            metadata["posted_to"] = []

        # Ensure platforms field exists (defines where to post)
        if "platforms" not in metadata:
            metadata["platforms"] = ["twitter", "linkedin", "telegram"]

        return metadata
    except Exception as e:
        print(f"Error parsing {file_path}: {e}", file=sys.stderr)
        return {}


def get_platforms_to_post(metadata: Dict) -> List[str]:
    """
    Determine which platforms still need posting.
    Returns list of platforms that haven't been posted to yet.
    """
    all_platforms = set(metadata.get("platforms", ["twitter", "linkedin", "telegram"]))
    posted_platforms = set(metadata.get("posted_to", []))

    # Return platforms that need posting
    return list(all_platforms - posted_platforms)


def update_frontmatter(file_path: str, platform: str) -> bool:
    """
    Update the markdown file to mark a platform as posted.
    Adds platform to 'posted_to' list in frontmatter.
    """
    try:
        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read()

        parts = content.split("---", 2)
        if len(parts) < 3:
            return False

        frontmatter = parts[1].strip()
        metadata = yaml.safe_load(frontmatter)

        # Add platform to posted_to list
        if "posted_to" not in metadata:
            metadata["posted_to"] = []

        if platform not in metadata["posted_to"]:
            metadata["posted_to"].append(platform)

        # Add timestamp for this posting
        timestamp_key = f"posted_{platform}_at"
        metadata[timestamp_key] = datetime.utcnow().isoformat()

        # Reconstruct the file
        new_frontmatter = yaml.dump(
            metadata, default_flow_style=False, allow_unicode=True
        )
        new_content = f"---\n{new_frontmatter}---{parts[2]}"

        with open(file_path, "w", encoding="utf-8") as f:
            f.write(new_content)

        return True
    except Exception as e:
        print(f"Error updating {file_path}: {e}", file=sys.stderr)
        return False


def commit_frontmatter_changes(files: List[str], message: str = None):
    """
    Commit the updated markdown files back to git.
    """
    if not files:
        return

    try:
        # Stage the files
        subprocess.run(["git", "add"] + files, check=True)

        # Create commit message
        if message is None:
            message = f"Update posting status for {len(files)} blog post(s) [skip ci]"

        # Commit
        subprocess.run(["git", "commit", "-m", message], check=True)

        print(f"✓ Committed updates to {len(files)} file(s)")
    except subprocess.CalledProcessError as e:
        print(f"Error committing changes: {e}", file=sys.stderr)


def detect_posts_to_publish() -> Dict[str, Dict]:
    """
    Main function: Detect which posts need publishing to which platforms.

    Returns:
        Dictionary mapping file paths to posting information:
        {
            'blog_posts/example.md': {
                'metadata': {...},
                'platforms': ['twitter', 'linkedin'],
                'url': 'https://mlship.dev/blog/example'
            }
        }
    """
    posts_to_publish = {}

    # Get changed files from last commit
    changed_files = get_changed_markdown_files()

    if not changed_files:
        # No error message here - just return empty dict
        return posts_to_publish

    # Only print to stderr so it doesn't break JSON output
    print(f"Found {len(changed_files)} changed markdown file(s)", file=sys.stderr)

    for file_path in changed_files:
        metadata = parse_frontmatter(file_path)

        if not metadata:
            continue

        # Get platforms that still need posting
        platforms = get_platforms_to_post(metadata)

        if platforms:
            # Build the full URL
            slug = metadata.get("slug", "")
            url = f"https://mlship.dev/blog/{slug}"

            posts_to_publish[file_path] = {
                "metadata": metadata,
                "platforms": platforms,
                "url": url,
                "title": metadata.get("title", ""),
                "abstract": metadata.get("abstract", ""),
                "excerpt": metadata.get("excerpt", ""),
            }

            print(
                f"✓ {file_path} needs posting to: {', '.join(platforms)}",
                file=sys.stderr,
            )
        else:
            print(f"○ {file_path} already posted to all platforms", file=sys.stderr)

    return posts_to_publish


def main():
    """
    CLI entry point. Can be used in two modes:
    1. Detect mode (default): Returns JSON of posts to publish
    2. Update mode: Marks a post as published to a platform
    """
    import json
    import argparse

    parser = argparse.ArgumentParser(
        description="Detect blog posts that need cross-posting"
    )
    parser.add_argument(
        "--update",
        metavar="FILE:PLATFORM",
        help="Update a file to mark platform as posted (e.g., blog_posts/example.md:twitter)",
    )
    parser.add_argument(
        "--commit", action="store_true", help="Commit the changes after updating"
    )

    args = parser.parse_args()

    if args.update:
        # Update mode: mark a post as published
        try:
            file_path, platform = args.update.split(":")
            success = update_frontmatter(file_path, platform)

            if success:
                print(f"✓ Marked {file_path} as posted to {platform}")

                if args.commit:
                    commit_frontmatter_changes(
                        [file_path],
                        f"Mark {file_path} as posted to {platform} [skip ci]",
                    )
            else:
                print(f"✗ Failed to update {file_path}", file=sys.stderr)
                sys.exit(1)
        except ValueError:
            print("Error: --update format should be FILE:PLATFORM", file=sys.stderr)
            sys.exit(1)
    else:
        # Detect mode: find posts that need publishing
        posts = detect_posts_to_publish()

        # Always output valid JSON to stdout (even if empty)
        print(json.dumps(posts, indent=2))

        # Log to stderr so it doesn't interfere with JSON
        if posts:
            print(
                f"\nDetected {len(posts)} post(s) needing publication", file=sys.stderr
            )
        else:
            print("No posts need publishing", file=sys.stderr)

        # Exit with 0 always - let the workflow check if dict is empty
        sys.exit(0)


if __name__ == "__main__":
    main()
