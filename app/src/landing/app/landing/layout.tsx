import type { Metadata } from "next";

export const metadata: Metadata = {
	title: "Thanos",
	description:
		"Open-source platform for running parallel AI coding agents. Spawn Claude Code, Codex, Aider, and more in isolated worktrees — all managed from one dashboard.",
	openGraph: {
		type: "website",
		url: "https://tinhtran.dev/landing",
		siteName: "Thanos",
		title: "Thanos",
		description:
			"Open-source platform for running parallel AI coding agents. Spawn Claude Code, Codex, Aider, and more in isolated worktrees — all managed from one dashboard.",
		images: [{ url: "/og-image.png", width: 1024, height: 1024, alt: "Thanos" }],
	},
	twitter: {
		card: "summary",
		site: "@tinhtran",
		creator: "@tinhtran",
		title: "Thanos",
		description:
			"Open-source platform for running parallel AI coding agents. Spawn Claude Code, Codex, Aider, and more in isolated worktrees — all managed from one dashboard.",
		images: ["/og-image.png"],
	},
	alternates: {
		canonical: "https://tinhtran.dev/",
	},
};

export default function LandingLayout({ children }: { children: React.ReactNode }) {
	return <>{children}</>;
}
