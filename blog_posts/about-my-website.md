---
abstract: How I build my website with Claude, placeholder blogpost
date: "2024-08-13"
excerpt: This is my experience building a website with Claude AI
platforms:
  - telegram
  - twitter
  - mastodon
  - devto
posted_mastodon_at: "2025-10-24T17:18:10.393815"
posted_telegram_at: "2025-10-23T20:00:45.001414"
posted_to:
  - telegram
  - twitter
  - mastodon
posted_twitter_at: "2025-10-24T05:43:43.823135"
slug: my-website
title: How I built my website with Claude
---

# Creating My Personal Website with Claude AI: A Journey of Discovery

_[View the website code on GitHub](https://github.com/thatbagu/cv)_

## Introduction

I decided to completely recreate my personal website. The previous version was built with Next.js, but as someone who's not a frontend developer, I found myself out of my depth. I had simply forked another developer's code and made minor personalizations. This time, I wanted a stack I could understand and maintain, so I opted for HTMX and FastAPI.

## The Early Stages

everything was simple and "macaroni coded":

- The entire backend resided in a single file
- CSS was minimal
- HTML was split into pages

Claude AI understood the context well at this stage, but challenges arose as the project grew.

## Challenges with Claude AI

As the project expanded, several issues emerged:

1. **Code Duplication**: Claude AI often produced redundant code.
2. **Debugging Difficulties**: As the codebase grew, identifying and fixing bugs became more challenging.
3. **Context Limitations**: Feeding the entire codebase to Claude AI became inefficient.

## Refining the Approach

To overcome these challenges, I adopted a new strategy:

1. **Focused Problem-Solving**: I began feeding Claude AI only the problematic areas of the code.
2. **Providing Context**: Instead of sharing the entire codebase, I explained the context myself.
3. **Directing the AI**: I had to maintain a big-picture view and guide Claude AI to:
   - Split files appropriately
   - Refine global logic
   - Adhere to the chosen technology stack

This approach helped prevent ending up with a single-file "spaghetti code" mess.

## Tackling Specific Areas

### Cloud Infrastructure and CI/CD

One of the most challenging aspects was setting up CI/CD with cloud services:

- Claude AI struggled with debugging and guiding me through cloud setup.
- However, it successfully generated working OpenTofu IaC files for my Google Cloud Platform setup.

### Frontend Design

Claude AI excelled in frontend tasks:

1. **CSS Configuration**: It consistently produced near-perfect results based on my descriptions of desired layouts and styles.
2. **Icons and ASCII Art**: After multiple attempts and specific directions, Claude AI managed to create satisfactory, if not perfect, designs.

### Backend Development

Backend work proved more challenging:

- Claude AI took longer to understand and implement backend logic.
- Bug occurrences were frequent.
- Error resolution was aided by sharing terminal outputs, which Claude AI could usually interpret correctly.

## Conclusion

Using Claude AI significantly accelerated my website development process. However, it's important to note:

1. The AI performed better with frontend tasks compared to backend development.
2. Human oversight and direction were crucial throughout the process.
3. Post-development code refinement was necessary to ensure optimal performance and maintainability.

In summary, while Claude AI proved to be a valuable tool in expediting development, it's essential to approach AI-assisted coding with a critical eye and be prepared to refine the output.

## A Meta Moment: Refining This Blog Post with Claude AI

Interestingly, the process of creating this blog post mirrors the website development experience I've just described. I initially wrote a rough draft of my thoughts and experiences, then turned to Claude AI for assistance in refining and structuring the content.
