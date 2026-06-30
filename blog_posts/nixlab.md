---
abstract: A NixOS homelab template for a self-hosted k3s cluster, built out of frustration with minimal setups and a belief that owning your compute matters
date: "2026-06-30"
excerpt: >-
  I started this after watching a Dreams of Autonomy video in 2024, bought Beelink mini PCs, fought through Kazakhstani customs, and ended up deeply unsatisfied with minimal NixOS homelab setups that only bootstrapped the cluster but left service deployment out. So I built my own: a fully declarative k3s cluster where Colmena handles the system layer and a NixOS activation script applies Helm charts on top. The entire stack (MetalLB, Longhorn, Pi-hole, WireGuard, Nextcloud, cert-manager) lives in a single vars.nix. Every boot wipes / via btrfs rollback so nodes are always in a known-good state. The deeper reason: I think owning your compute is becoming essential as corporations buy more and more for datacenters.
platforms:
  - telegram
  - twitter
  - mastodon
  - devto
posted_to: []
slug: nixlab
title: "nixlab: Own Your Compute"
---

_[GitHub](https://github.com/thatbagu/nixlab)_

![nixlab](/assets/projects/nixlab.jpeg)

## How it started

This project has been cooking for a while. It started with a [Dreams of Autonomy video](https://www.youtube.com/watch?v=2yplBzPCghA) back in 2024 that sent me down the homelab rabbit hole. I bought a pair of Beelink mini PCs and set out to replicate the setup, but the result left me deeply unsatisfied. NixOS was used only to configure the initial cluster bootstrap, not for actual service deployment. That was not what I signed up for.

Getting the hardware into Kazakhstan was kinda difficult. I had to contact customs, pay import taxes, fill out a couple of forms. Shout out to Beelink for helping out and covering shipping costs.

## Building around constraints

With the hardware on the desk, I started working within the constraints of what I already had. The approach I landed on: custom systemd services that apply Helm charts right after Colmena pushes a NixOS config update to a node. It felt clunky at first, but it grew on me. Eventually it became natural. Colmena handles the declarative system layer, the activation scripts handle the Kubernetes layer. One `colmena apply` and the whole stack converges.

The full cluster (MetalLB, Longhorn, nginx ingress, Pi-hole, cert-manager, WireGuard VPN, Nextcloud, and more) is declared in a single `vars.nix` file. Adding a node means copying a hardware config template, filling in an IP and a disk, and applying. Everything else is derived automatically.

Every boot wipes `/` via a btrfs rollback in the initrd. Only `/persist` survives. Nodes are always in a known-good state, and state that matters is explicitly declared.

## No kubectl apply by hand

Helm charts are never applied manually. When you run `colmena apply` or `nixos-rebuild switch`, a NixOS activation script fires before the system even finishes switching. It takes every service definition, which are Nix attrsets rendered to YAML at build time via `nixhelm` and `nix-kube-generators` and baked into the Nix store, writes them to `/var/lib/kubernetes/manifests/`, and restarts a `k8s-deploy` systemd oneshot service.

That service polls the Kubernetes API until it responds, creates all required namespaces, then walks through the deployment groups in dependency order: core infrastructure first, then networking, DNS, TLS, VPN, and finally apps. Each group writes a sentinel file when it succeeds, so a partial failure reruns only what did not finish. The deploy is a side effect of a system rebuild.

## Isolating it from my dotfiles

The homelab config lives inside my larger dotfiles repo, shared with everything I use daily. To publish it, I had to strip it down and isolate only the homelab parts so anyone could clone it and replicate the setup without needing the rest of my personal config. That cleanup turned out to be useful on its own: it forced me to make the template actually self-contained.

## Why own a homelab

The practical reasons are obvious: self-hosted storage, ad blocking, a VPN, full control over your services. But for me the larger reason is ideological.

Compute is becoming as essential to daily life as food and water. The more people outsource that to large corporations, surrendering their data, their workflows, their dependencies, the more fragile and unfree we become as a society. Owning your compute, even on a small scale, is a meaningful act. It keeps you literate about the infrastructure that runs your life, and it keeps that infrastructure under your control.

That is probably a lot to read into a pair of Beelink mini PCs. But I think this kind of effort is part of a larger mission, and I hope nixlab makes it a little easier for someone else to start.
