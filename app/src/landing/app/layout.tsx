import type { Metadata } from "next";
import "../styles/globals.css";

export const metadata: Metadata = {
	title: "Maestro",
	description:
		"An AI Engineering Workbench that orchestrates AI agents through a structured, human-controlled software development workflow.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
	return (
		<html lang="en" suppressHydrationWarning>
			<body>{children}</body>
		</html>
	);
}
