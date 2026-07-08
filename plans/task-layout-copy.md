# Task Plan: Copy layout Wallfacer Production Behavior Into Thanos

## Objective

Build Thanos into a Wails-native, local-first AI development workbench with the
production behaviors represented in the `wallfacer/` reference project, while
keeping Thanos implementation independent and Thanos-native.

Wallfacer is a reference for product behavior and architecture shape only. Do
not copy implementation code.
## Product Scope

Thanos app/frontend must same function and structure with ưallfacer/frontend: 
- Hook
- Structure project 
- i18n
- Layout
- styles
- view 
- Implement app logo
- Stores

## Implementation Status

- [x] Hook structure added through `src/hooks/useWorkspaceController.ts`.
- [x] Project structure aligned with `hooks`, `stores`, `i18n`, `views`, and modular `styles` directories.
- [x] i18n foundation added with keyed English copy and interpolation tests.
- [x] Layout updated to match the Wallfacer local workbench shell shape: grouped workspace and inspect navigation, workspace switcher, status metadata, and Thanos branding.
- [x] App logo wired into the shell brand mark.
- [x] Store helpers added for workspace defaults and refresh identity preservation.
- [x] Tests added for navigation shape, i18n interpolation, and workspace store behavior.
