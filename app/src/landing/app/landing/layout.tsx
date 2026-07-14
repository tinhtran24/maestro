import type { Metadata } from "next";

export const metadata: Metadata = {
	title: "Maestro",
	description:
		"An AI Engineering Workbench that orchestrates AI agents through a structured, human-controlled software development workflow. You conduct; agents like Claude Code, Codex, and Gemini execute tasks in isolated worktrees — all from one dashboard.",
	openGraph: {
		type: "website",
		url: "https://tinhtran.dev/landing",
		siteName: "Maestro",
		title: "Maestro",
		description:
			"An AI Engineering Workbench that orchestrates AI agents through a structured, human-controlled software development workflow. You conduct; agents like Claude Code, Codex, and Gemini execute tasks in isolated worktrees — all from one dashboard.",
		images: [{ url: "/og-image.png", width: 1024, height: 1024, alt: "Maestro" }],
	},
	twitter: {
		card: "summary",
		site: "@tinhtran",
		creator: "@tinhtran",
		title: "Maestro",
		description:
			"An AI Engineering Workbench that orchestrates AI agents through a structured, human-controlled software development workflow. You conduct; agents like Claude Code, Codex, and Gemini execute tasks in isolated worktrees — all from one dashboard.",
		images: ["/og-image.png"],
	},
	alternates: {
		canonical: "https://tinhtran.dev/",
	},
};

export default function LandingLayout({ children }: { children: React.ReactNode }) {
	return <>{children}</>;
}
