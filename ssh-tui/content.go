package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// osc8 wraps text in an OSC 8 hyperlink (clickable in Kitty, WezTerm, iTerm2, etc.)
func osc8(url, text string) string {
	return fmt.Sprintf("\033]8;;%s\a%s\033]8;;\a", url, text)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func fmtH1(text string, st styles) string {
	bar := strings.Repeat("━", utf8.RuneCountInString(text)+4)
	return "\n" + st.h1.Render(bar) + "\n" +
		st.h1.Render("  "+text) + "\n" +
		st.h1.Render(bar) + "\n"
}

func fmtH2(text string, st styles) string {
	under := st.lav.Render(strings.Repeat("─", utf8.RuneCountInString(text)+3))
	return "\n" + st.h2.Render("▌ "+text) + "\n" + under + "\n"
}

func fmtH3(text string, st styles) string {
	return "\n" + st.h3.Render("  ◈ "+text) + "\n"
}

func fmtPara(text string, w int, st styles) string {
	return st.body.Width(w-4).PaddingLeft(2).Render(text) + "\n\n"
}

func fmtBullets(items []string, w int, st styles) string {
	bullet := st.peach.Render("•")
	var sb strings.Builder
	for _, item := range items {
		wrapped := st.body.Width(w - 6).Render(item)
		lines := strings.Split(wrapped, "\n")
		for i, line := range lines {
			line = strings.TrimRight(line, " ")
			if line == "" {
				continue
			}
			if i == 0 {
				sb.WriteString("  " + bullet + " " + line + "\n")
			} else {
				sb.WriteString("    " + line + "\n")
			}
		}
	}
	sb.WriteString("\n")
	return sb.String()
}

func fmtLink(label, url string, st styles) string {
	arrow := osc8(url, st.sap.Render("→ "+label))
	urlText := st.sub.Faint(true).Render(url)
	return "  " + arrow + "  " + urlText + "\n\n"
}

// ── Home ──────────────────────────────────────────────────────────────────────

func renderHome(w int, st styles, frames []string, frameIdx int, steppe *steppeScene,
	mPosts []MastodonPost, loading bool,
	searching bool, searchView, searchQuery string,
	searchResults []searchResult, searchCursor int) string {

	var sb strings.Builder

	// Steppe landscape scene
	sb.WriteString(renderSteppeScene(steppe, st))
	sb.WriteString("\n")

	// Title box
	const inner = "Egor Kosaretsky — MLOps Engineer"
	boxW := utf8.RuneCountInString(inner) + 4
	sb.WriteString(st.h1.Render("  ╔"+strings.Repeat("═", boxW)+"╗") + "\n")
	sb.WriteString(st.h1.Render("  ║  "+inner+"  ║") + "\n")
	sb.WriteString(st.h1.Render("  ╚"+strings.Repeat("═", boxW)+"╝") + "\n\n")

	sb.WriteString(st.body.Render("  MLOps Engineer · Almaty, Kazakhstan · 4+ yrs ML platform") + "\n\n")

	// Navigation hints
	sb.WriteString(st.lav.Render("  Navigate:") + "\n\n")
	sb.WriteString("  " + st.peach.Render("[2]") + " " + st.body.Render("About Me & Full CV") + "\n")
	sb.WriteString("  " + st.peach.Render("[3]") + " " + st.body.Render("Projects") + "\n")
	sb.WriteString("  " + st.peach.Render("[4]") + " " + st.body.Render("Blog") + "\n")
	sb.WriteString("  " + st.peach.Render("[5]") + " " + st.body.Render("Contact") + "\n\n")
	sb.WriteString("  " + st.sub.Render("Web:    ") + osc8("https://mlship.dev", st.sap.Render("https://mlship.dev")) + "\n")
	sb.WriteString("  " + st.sub.Render("GitHub: ") + osc8("https://github.com/thatbagu", st.sap.Render("https://github.com/thatbagu")) + "\n\n")

	// Search widget
	sb.WriteString(renderSearchWidget(w, st, searching, searchView, searchQuery, searchResults, searchCursor))

	// Mastodon posts
	sb.WriteString(fmtH2("Recent Posts", st))
	if loading {
		sb.WriteString(st.sub.Render("  Loading…") + "\n\n")
	} else if len(mPosts) == 0 {
		sb.WriteString(st.sub.Render("  No recent posts found.") + "\n\n")
	} else {
		for _, p := range mPosts {
			sb.WriteString(st.sub.Render("  "+p.Date) + "\n")
			sb.WriteString(st.body.Width(w-4).PaddingLeft(2).Render(p.Text) + "\n")
			if p.URL != "" {
				sb.WriteString("  " + osc8(p.URL, st.sap.Faint(true).Render("↗ mastodon")) + "\n")
			}
			sb.WriteString("\n")
		}
	}

	// ASCII cat animation (full size, centered)
	if len(frames) > 0 {
		sb.WriteString(renderFrame(frames[frameIdx], w, st))
	}

	return sb.String()
}

func renderSearchWidget(w int, st styles, searching bool, searchView, searchQuery string,
	results []searchResult, cursor int) string {

	var sb strings.Builder
	if !searching {
		sb.WriteString("  " + st.sub.Faint(true).Render("/: search across posts & projects") + "\n\n")
		return sb.String()
	}

	// Box width: leave 4 chars for margin on each side
	boxInner := max(30, w-8)
	top := "╭── Search " + strings.Repeat("─", max(0, boxInner-10)) + "╮"
	bot := "╰" + strings.Repeat("─", boxInner) + "╯"

	sb.WriteString("  " + st.lav.Render(top) + "\n")
	sb.WriteString("  " + st.lav.Render("│") + " " + searchView + " " + st.lav.Render("│") + "\n")
	sb.WriteString("  " + st.lav.Render(bot) + "\n\n")

	if len(results) > 0 {
		for i, r := range results {
			if i == cursor {
				sb.WriteString("  " + st.peach.Bold(true).Render("▶ "+r.label) + "\n")
			} else {
				sb.WriteString("  " + st.sub.Render("  "+r.label) + "\n")
			}
		}
	} else if strings.TrimSpace(searchQuery) != "" {
		sb.WriteString("  " + st.sub.Render("  no results") + "\n")
	}
	sb.WriteString("\n")
	return sb.String()
}

// renderFrame renders the ASCII cat frame at full resolution, centered.
func renderFrame(frame string, w int, st styles) string {
	lines := strings.Split(strings.TrimRight(frame, "\n"), "\n")
	maxW := 0
	for _, line := range lines {
		if n := utf8.RuneCountInString(strings.TrimRight(line, " ")); n > maxW {
			maxW = n
		}
	}
	leftPad := strings.Repeat(" ", max(0, (w-maxW)/2))
	var sb strings.Builder
	for _, line := range lines {
		sb.WriteString(leftPad + st.peach.Render(strings.TrimRight(line, " ")) + "\n")
	}
	return sb.String()
}

// ── About ─────────────────────────────────────────────────────────────────────

func renderAbout(w int, st styles) string {
	var sb strings.Builder

	sb.WriteString(fmtH1("Egor Kosaretsky", st))
	sb.WriteString(st.h3.Render("  MLOps Engineer | Almaty, Kazakhstan") + "\n\n")

	// Profile photo centered above summary
	const profileW = 28
	if img := centeredHalfBlock(profileImg, profileW, w); img != "" {
		sb.WriteString(img)
		sb.WriteString("\n")
	}

	sb.WriteString(fmtH2("Summary", st))
	sb.WriteString(fmtPara(
		"MLOps engineer with 4+ years of experience specializing in ML platform infrastructure at scale. "+
			"ITPEC FE certified. Seeking relocation to Japan — no work restrictions.",
		w, st))

	sb.WriteString(fmtH2("Skills", st))
	sb.WriteString(fmtBullets([]string{
		"ML Infrastructure: Kubeflow, Ray, Nvidia Triton, MLflow, ClearML, Apache Airflow, DVC, Apache Spark",
		"Cloud & DevOps: GCP, AWS, SageMaker, Databricks, Docker, Kubernetes, Terraform, Terragrunt, Nix, Prometheus, Grafana",
		"Programming: Python, PyTorch, Lightning, FastAPI, Pandas, Polars, scikit-learn, NumPy, Snakemake, Pytest, SQLAlchemy, R, Bash",
		"Machine Learning: Model serving, Inference optimization, Federated Learning, Gradient Boosting, Platform engineering",
	}, w, st))

	sb.WriteString(fmtH2("Professional Experience", st))

	sb.WriteString(fmtH3("inDrive | MLOps Engineer", st))
	sb.WriteString(st.sub.Render("  Kazakhstan, Almaty | 06.2024 – present") + "\n")
	sb.WriteString(fmtBullets([]string{
		"Built and maintained the Kubeflow/GCP ML platform and Ray inference layer for data science teams across a 600-engineer org",
		"Designed and shipped Inflow SDK — internal Python SDK standardizing the ML lifecycle from local dev to SageMaker, used in production by business services",
		"Implemented CI inference testing (testcontainers), preset model packaging, and OIDC client — cut model time-to-production by ~30%",
		"Led Databricks rollout (AI Platform 2.0): multi-region workspaces, GCP isolation, BigQuery/Lakehouse Federation, compute migration",
	}, w, st))

	sb.WriteString(fmtH3("Third Opinion AI | MLOps Engineer", st))
	sb.WriteString(st.sub.Render("  Russia, Moscow (remote) | 09.2023 – 06.2024") + "\n")
	sb.WriteString(fmtBullets([]string{
		"Co-designed ML platform architecture with the Head of AI; built a custom model storage and deployment interface on Nvidia Triton",
		"Optimized Ray inference pipelines and resolved critical cluster instability — met accuracy and latency SLAs enabling government contract wins",
		"Migrated CV research workflows to ClearML; implemented CI/CD pipelines and Kubeflow tooling for the research team",
	}, w, st))

	sb.WriteString(fmtH3("DevSect | ML Engineer", st))
	sb.WriteString(st.sub.Render("  Russia, remote | 04.2023 – 09.2023") + "\n")
	sb.WriteString(fmtBullets([]string{
		"Implemented ML microservices (FastAPI, Docker, Kubernetes) for facial recognition, Text-to-Speech, and Speech-to-Text",
		"Achieved a 12% increase in user app retention by integrating ML microservices into the mobile app",
	}, w, st))

	sb.WriteString(fmtH3("GENXT | Bioinformatician / Data Scientist", st))
	sb.WriteString(st.sub.Render("  UK, remote | 09.2021 – 04.2023") + "\n")
	sb.WriteString(fmtBullets([]string{
		"Built parallel Snakemake genomic data pipelines; EDA/PCA with Seaborn/Plotly; fine-tuned CatBoost/XGBoost models",
		"Shipped federated learning proof-of-concept (PyTorch, R) — adopted into a partner genetic company's production pipeline",
	}, w, st))

	sb.WriteString(fmtH2("Education", st))
	sb.WriteString(fmtH3("Lomonosov Moscow State University | Bioinformatics", st))
	sb.WriteString(st.sub.Render("  Russia, Moscow | 2018 – 2023") + "\n")
	sb.WriteString(fmtBullets([]string{
		"Focus on genomics and data analysis; coursework in Probability Theory, Statistics with R, and Python",
		"Additional training: ML for Genomics at Max Delbrück Center for Molecular Medicine",
	}, w, st))

	sb.WriteString(fmtH2("Certifications", st))
	sb.WriteString(fmtBullets([]string{
		"ITPEC Fundamental Information Technology Engineer Examination (FE) — Passed April 2026. Recognized by Japan Ministry of Justice for Engineer visa qualification.",
	}, w, st))

	sb.WriteString(fmtH2("Languages", st))
	sb.WriteString(fmtBullets([]string{
		"Russian: Native",
		"English: Fluent",
		"German: Business level",
		"Japanese: Elementary — JLPT N5 certified, studying toward N4",
	}, w, st))

	sb.WriteString(fmtH2("Papers & Talks", st))
	sb.WriteString(fmtBullets([]string{
		"Efficacy of Federated Learning on Genomic Data — federated learning engine proof of concept for genetic (SNP) data",
		"GRAPE: Genomic Relatedness Detection Pipeline — open-source pipeline for large-scale relative search in genomic datasets",
		"Growing an ML Platform from Scratch to Enterprise — Kolesa Conf, October 2025",
		"Why (not) to choose Kubeflow as your ML platform — conference talk, January 2025",
		"MLOps Skills as a Competitive Advantage for Data Analysts — Data Community Birthday, Almaty, February 2026",
	}, w, st))

	return sb.String()
}

// ── Projects ──────────────────────────────────────────────────────────────────

func renderProjects(w int, st styles) string {
	const projImgW = 36
	var sb strings.Builder
	sb.WriteString(fmtH2("Projects", st))
	for _, p := range allProjects {
		sb.WriteString(fmtH3(p.name, st))
		if img, ok := projectImgs[p.imgKey]; ok {
			sb.WriteString(centeredHalfBlock(img, projImgW, w))
			sb.WriteString("\n")
		}
		sb.WriteString(fmtPara(p.desc, w, st))
		sb.WriteString("  " + st.sub.Render("Tech: ") + st.body.Render(p.tech) + "\n\n")
		sb.WriteString(fmtLink("View", p.url, st))
	}
	return sb.String()
}

// ── Contact ───────────────────────────────────────────────────────────────────

func renderContact(w int, st styles) string {
	var sb strings.Builder
	sb.WriteString(fmtH2("Contact", st))
	sb.WriteString(fmtPara(
		"I'm always open to new opportunities and collaborations. Feel free to reach out!",
		w, st))

	type entry struct{ label, val, url string }
	for _, e := range []entry{
		{"Email:    ", "egor@mlship.dev", "mailto:egor@mlship.dev"},
		{"LinkedIn: ", "linkedin.com/in/egor-kosaretskiy", "https://www.linkedin.com/in/egor-kosaretskiy/"},
		{"GitHub:   ", "github.com/thatbagu", "https://github.com/thatbagu"},
		{"Web:      ", "https://mlship.dev", "https://mlship.dev"},
		{"PGP Key:  ", "mlship.dev/pgp  (A9B2618B04CED76F)", "https://mlship.dev/pgp"},
	} {
		sb.WriteString("  " + st.sub.Render(e.label) + osc8(e.url, st.sap.Render(e.val)) + "\n")
	}
	sb.WriteString("\n")
	sb.WriteString(fmtPara(
		"Available for MLOps/ML platform engineering roles, consulting, and open-source collaboration.",
		w, st))
	return sb.String()
}
