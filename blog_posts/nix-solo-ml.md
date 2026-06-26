---
abstract: Quick start cloud ML development environment with Nix
date: "2026-06-27"
excerpt: Automatic ML setup through Nix on AWS with basic logging services and model deployment
platforms:
  - telegram
  - twitter
  - mastodon
  - devto
slug: nix-solo-ml
title: High velocity ML development with Nix and AWS
---

_[GitHub](https://github.com/thatbagu/nix-ml-solo) | [Download slides if you have been to my talk](/assets/blog/nix-ml-solo-talk.pdf)_

## The need for speed

Over the past several years I have developed severe ADHD by consuming short-form content. Instead of fixing it by exercising proper habits, meditating, or whatever, I decided to lean into it. I read somewhere that ADHD also comes with superpowers: it is possible to catch a flow state (hyperfocus) where you can concentrate on something long enough to actually get things done. But this flow is fragile. It needs a constant feeling of speed and forward momentum.

Preserving that state pushed me to eliminate every break and barrier between a thought and actual code running on a machine. A couple of years ago I kept catching myself distracted and unable to write code even when my head was already running, so I invested time, money, and skill into a split keyboard and blind typing on the Miryoku layout. Then it bothered me that navigating an IDE was slow and forced me to keep track of where everything lives and how to configure and run things, so I built my own editor with Neovim. Even then, developing on common Linux distros like Ubuntu felt inefficient: installed packages quickly polluted the system, projects conflicted over different dependency versions, and there was no guarantee you could reproduce your environment if you came back to a project months later. That is when I discovered Nix and rebuilt my system, declarative, reproducible, and easy to manage. Local development became nearly frictionless. The obvious next step was bringing that same fluidity to the cloud.

## The idea: unified environments

I finally decided it was time to make Nix work on the cloud for my ML applications. Nix is built around reproducibility and seamless environments that are quick to spin up, and that property should hold equally well on a remote instance as on a local machine. This is the central idea behind nix-ml-solo: a fully unified environment between any cloud resource and any local machine, so a data scientist never has to think about which machine they are developing on.

`devenv.nix` is the single source of truth. Change values here and they flow everywhere: Terraform resource names, S3 buckets, ports.

```nix
let
  project       = "nix-ml-solo";   # all AWS resource names derive from this
  environment   = "dev";
  mlflowPort    = 5000;
  jupyterPort   = 8888;
  inferencePort = 5001;
in
{ ... }
```

Switching from local development to a full cloud setup is one line:

```nix
# env.INFRA_MODE = "cloud";  # uncomment this
```

In local mode everything runs on your machine and AWS costs are essentially zero (S3 only). Flip to cloud and the same environment materializes on an EC2 NixOS VM and inside SageMaker containers, bit-for-bit identical, because they all share the same Nix closure. MLflow runs on EC2 and is accessed via SSH tunnel, so your tracking URI stays `localhost:5000` regardless of where the training job actually runs. DVC handles data through an S3 remote that both your laptop and SageMaker pull from directly.

## Architecture

![nix-ml-solo architecture](/assets/blog/nix-ml-solo-architecture.png)

## The stack

The setup is intentionally minimal but fully viable: MLflow + DVC + SageMaker. Experiment, log, deploy. I think this is the core MVP for any ML workflow. Pipeline orchestration, feature engineering, and data ops are on the roadmap. The project is open source at [github.com/thatbagu/nix-ml-solo](https://github.com/thatbagu/nix-ml-solo).
