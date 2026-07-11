---
abstract: While everyone builds visual dashboards for AI agent orchestration, I went
  the opposite direction. Here is why a terminal, home row mods, and Zellij is faster
  than any GUI for managing multiple AI agent sessions.
date: '2026-07-11'
excerpt: Everyone is talking about how to orchestrate AI agents. Nobody is talking
  about the actual UX of driving them fast. The industry default is visual dashboards,
  reasoning panels, and orchestration UIs. I went the opposite direction and my workspace
  is a terminal. This post covers the two tools that made that setup genuinely faster
  than anything I have tried with a GUI.
platforms:
- telegram
- twitter
- mastodon
- devto
slug: terminal-ai-workspace
title: 'Optimizing the Pilot: My Terminal-First AI Agent Workspace'
---

## The missing layer

Everyone is talking about how AI agents have changed their workflow, how to properly orchestrate them, which frameworks to use, how to chain them together. That conversation is useful. But there is a different layer that nobody seems to be discussing: the actual UX of fast, fluid interaction with your agents while they are running, whether you are managing one or ten.

The industry's answer in 2026 is visual dashboards, orchestration UIs, and reasoning panels. Beautiful, well-designed tools that show you what the agent is doing, surface its reasoning, and let you intervene at each step. I tried several of them. They are well-built. I still went the opposite direction.

My workspace is a terminal. And it is faster than anything I have tried with a GUI.

This is not about being contrarian. It is about what actually keeps you in flow when you are juggling multiple agentic sessions simultaneously. Two tools made that possible for me.

## Home row mods

The first and most important prerequisite is home row mods. If you are not familiar: your home row keys (A, S, D, F on the left and J, K, L, ; on the right) act as regular letters when tapped, but become modifier keys (Ctrl, Alt, Shift, Super) when held. On a Mac you set this up with [Karabiner-Elements](https://gregorias.github.io/posts/home-row-mods-karabiner-elements/).

This sounds like a minor ergonomic tweak. It is not. It eliminates the constant hand movement that breaks your flow. Instead of reaching to the corners of your keyboard every time you need a modifier, your fingers never leave the home row. The modifiers are already at your fingertips.

The reason this matters so much in an agentic workflow is that switching between panes, tabs, and sessions is something you do constantly. Every switch is a key combination. If every key combination requires moving your hand, the friction compounds quickly. With home row mods it does not. The interaction becomes almost unconscious.

This is the prerequisite to everything else. It enables truly multilayered, fluid interaction with a terminal multiplexer.

## Zellij

[Zellij](https://zellij.dev) is the king of terminal multiplexers. Some of you know tmux. Tmux is powerful but turning it into a proper workspace takes significant effort: you need a theme, a status bar plugin, sensible keybindings, session management config. It is a project on its own. Zellij works out of the box.

When you open Zellij for the first time it shows you the keybindings. There is nothing to configure before it is usable. From there you get:

- **Panes** split horizontally or vertically within a tab, each running an independent shell or process
- **Tabs** for grouping related panes, named and navigable with a single keypress
- **Sessions** that persist and can be attached and detached at will
- **Instant copy** by simply highlighting text — no extra keypress required

What makes this exceptional for AI agent workflows is that each Claude Code session runs in its own pane. You see all of them at once. You navigate between them with `Alt+h/j/k/l`. When an agent finishes and you want to give it a new task, you are one keypress away regardless of which pane you are currently in. When you want to fullscreen a single pane to read a long response, that is one keypress too. Everything is composable and immediate.

Configured with home row mods in mind, you can navigate through every layer of the multiplexer — tabs, panes, sessions — without moving your hands at all.

## What this looks like in practice

A typical session for me looks like this: three panes open, each running a separate Claude Code task. Left pane refactoring one module, top-right writing tests, bottom-right digging into a bug. I check in on each one, give direction where needed, and move on. The navigation between them takes less than a second and zero mental overhead.

This is the part the GUI tools have not solved. They are excellent at showing you what a single agent is doing. They are not designed for the physical experience of driving multiple agents in parallel at speed.

Most people are optimizing the agents. I am optimizing the pilot.

If you try this setup, start with home row mods first. It takes a week or two to stop accidentally triggering modifiers when you mean to type a letter, but once it clicks, going back feels like typing with oven mitts on. Then add Zellij. The combination changes how fast you can actually work.
