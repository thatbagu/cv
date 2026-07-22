package main

import "strings"

// projectDef holds metadata for a single project, shared between rendering and search.
type projectDef struct {
	name, imgKey, desc, tech, url string
}

// allProjects is the canonical list used by both renderProjects and doSearch.
var allProjects = []projectDef{
	{
		name:   "GRAPE",
		imgKey: "grape",
		desc:   "Open-source end-to-end Genomic RelAtedness detection PipelinE for large-scale relative search in genomic datasets.",
		tech:   "Snakemake · Docker · Python · Azure",
		url:    "https://github.com/genxnetwork/grape",
	},
	{
		name:   "Federated Learning — UK Biobank",
		imgKey: "federated",
		desc:   "Paper on proof-of-concept applications of Federated Learning on SNP data. Adopted by a partner genetic company in their production pipeline.",
		tech:   "Flower · PyTorch Lightning · scikit-learn · MLflow · Python",
		url:    "https://www.frontiersin.org/articles/10.3389/fdata.2024.1266031/full",
	},
	{
		name:   "nix-ml-solo",
		imgKey: "nix-ml-solo",
		desc:   "Solo ML stack on AWS with reproducible Nix environments, MLflow experiment tracking, DVC data versioning, and SageMaker training.",
		tech:   "Nix · Python · AWS SageMaker · MLflow · DVC · Shell",
		url:    "https://github.com/thatbagu/nix-ml-solo",
	},
	{
		name:   "nixlab",
		imgKey: "nixlab",
		desc:   "Opinionated NixOS flake for a self-hosted homelab on a multi-node k3s cluster. Nextcloud, Pi-hole, WireGuard, cert-manager declared from a single vars.nix. Every boot wipes / via btrfs rollback.",
		tech:   "NixOS · Nix · k3s · Kubernetes · WireGuard · Colmena · SOPS",
		url:    "https://github.com/thatbagu/nixlab",
	},
}

type searchResult struct {
	label   string
	section int
	blogIdx int // -1 for non-blog results
}

func doSearch(query string, posts []BlogPost) []searchResult {
	q := strings.TrimSpace(strings.ToLower(query))
	if q == "" {
		return nil
	}
	var results []searchResult
	for _, p := range allProjects {
		if strings.Contains(strings.ToLower(p.name), q) ||
			strings.Contains(strings.ToLower(p.desc), q) {
			results = append(results, searchResult{
				label:   "Project: " + p.name,
				section: 2,
				blogIdx: -1,
			})
		}
	}
	for i, p := range posts {
		if strings.Contains(strings.ToLower(p.Title), q) ||
			strings.Contains(strings.ToLower(p.Excerpt), q) ||
			strings.Contains(strings.ToLower(p.Body), q) {
			results = append(results, searchResult{
				label:   "Blog: " + p.Title,
				section: 3,
				blogIdx: i,
			})
		}
	}
	if len(results) > 6 {
		results = results[:6]
	}
	return results
}
