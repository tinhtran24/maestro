# Contributing

We love contributions! Join our community on Discord to get started.

## Join us on Discord

[![Discord](https://img.shields.io/badge/Discord-join%20the%20community-5865F2?style=for-the-badge&logo=discord&logoColor=white&logoSize=auto)](https://discord.com/invite/UZv7JjxbwG)

**Daily contributor sync:** Every day at **10:00 PM IST**

Get your issues verified by core contributors, ask questions, share progress, and learn from the community. New contributors are always welcome!

**Why join Discord?**

- Get your issues and PRs verified by core contributors before investing time
- Learn from experienced contributors in daily sync calls
- Share your progress and get feedback
- Get help troubleshooting in real-time
- Stay updated on the latest developments and roadmap

## Quick Start

1. **Join the Discord** - Connect with the community and get guidance
2. **Read the contributor contract** - See [AGENTS.md](AGENTS.md) for repo layout, daemon/API boundaries, and coding conventions
3. **Pick a focused problem** - Browse [open issues](https://github.com/AgentWrapper/maestro/issues) and choose one small enough for a focused PR
4. **Open a clear PR** - Keep changes narrow, explain user-visible impact, link issues, include tests
5. **Iterate with contributors** - Use review feedback to tighten the PR until verified

## Delivery protocol

Use this protocol for implementation work so changes are reviewable, traceable, and
safe to merge. The detailed command matrix and architecture boundaries live in
[AGENTS.md](AGENTS.md).

1. Create one branch per concern from `main`, named
   `<prefix>/<short-kebab-case-summary>`. Allowed prefixes are `feature/`, `bugfix/`,
   `hotfix/`, `refactor/`, `chore/`, `docs/`, and `test/`.
2. Keep the diff focused. Do not combine behavior changes with unrelated cleanup or
   broad refactors. Preserve backward compatibility unless the issue explicitly calls
   for a breaking change.
3. Make small, logical commits using Conventional Commits:
   `<type>(<optional-scope>): <imperative summary>`. For example,
   `fix(daemon): retain request id in error responses`. Use types such as `feat`,
   `fix`, `refactor`, `docs`, `test`, `chore`, `build`, `ci`, and `perf`.
4. Add or update tests for every behavior change. Update documentation when a public
   API, CLI behavior, setup instruction, or workflow changes.
5. Run the relevant formatting, linting, type checking, build, and test commands for
   the touched component. The pull-request template records what ran and why anything
   was not applicable.
6. Review the final diff and status. Do not commit secrets, local state, temporary
   files, build artifacts, or generated files unless the repository specifically
   requires a generated artifact (for example the API spec/types or sqlc output).

Use the pull-request template when opening a PR. Its summary, validation, branch, and
commit fields are the required delivery handoff.
