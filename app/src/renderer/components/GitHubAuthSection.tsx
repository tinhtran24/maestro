import { useQuery } from "@tanstack/react-query";
import { CheckCircle2, CircleAlert, Loader2, RefreshCw } from "lucide-react";
import type { ReactNode } from "react";
import { apiClient, apiErrorMessage } from "../lib/api-client";
import { Button } from "./ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";

export const githubAuthQueryKey = ["github-auth-status"] as const;

async function fetchGitHubAuth() {
	const { data, error } = await apiClient.GET("/api/v1/github/auth");
	if (error) throw new Error(apiErrorMessage(error, "Unable to check GitHub auth"));
	return data?.status;
}

export function GitHubAuthSection() {
	const query = useQuery({ queryKey: githubAuthQueryKey, queryFn: fetchGitHubAuth });
	const status = query.data;
	const healthy = Boolean(status?.authenticated);
	const Icon = query.isLoading ? Loader2 : healthy ? CheckCircle2 : CircleAlert;

	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-[13px]">GitHub</CardTitle>
			</CardHeader>
			<CardContent className="flex flex-col gap-4">
				<div className="flex items-start gap-3">
					<Icon className={`mt-0.5 size-4 ${query.isLoading ? "animate-spin text-muted-foreground" : healthy ? "text-success" : "text-warning"}`} />
					<div className="min-w-0 flex-1">
						<div className="text-[13px] font-medium text-foreground">
							{query.isLoading ? "Checking GitHub auth" : healthy ? "Authenticated" : "Action needed"}
						</div>
						<p className="mt-1 text-[12px] leading-5 text-muted-foreground">
							{query.isError
								? query.error instanceof Error
									? query.error.message
									: "Unable to check GitHub auth"
								: status?.message || "Thanos uses GitHub credentials for tracker intake, PRs, reviews, and CI status."}
						</p>
					</div>
				</div>

				<div className="flex flex-col gap-2 text-[12px]">
					<Row label="CLI">{status ? (status.installed ? status.binaryPath || "Installed" : "Missing") : "Unknown"}</Row>
					<Row label="Token">{status ? (status.authenticated ? status.source : "Not available") : "Unknown"}</Row>
				</div>

				{!healthy && status?.installCommand ? <CommandRow label="Install" command={status.installCommand} /> : null}
				{!healthy && status?.loginCommand ? <CommandRow label="Login" command={status.loginCommand} /> : null}

				<div>
					<Button type="button" variant="secondary" onClick={() => query.refetch()} disabled={query.isFetching}>
						{query.isFetching ? <Loader2 className="size-3.5 animate-spin" /> : <RefreshCw className="size-3.5" />}
						Refresh status
					</Button>
				</div>
			</CardContent>
		</Card>
	);
}

function Row({ label, children }: { label: string; children: ReactNode }) {
	return (
		<div className="flex items-center justify-between gap-3">
			<span className="text-muted-foreground">{label}</span>
			<span className="min-w-0 truncate text-right text-foreground">{children}</span>
		</div>
	);
}

function CommandRow({ label, command }: { label: string; command: string }) {
	return (
		<div className="rounded-md border border-border bg-surface/40 px-3 py-2 text-[12px]">
			<div className="mb-1 text-muted-foreground">{label}</div>
			<code className="text-foreground">{command}</code>
		</div>
	);
}
