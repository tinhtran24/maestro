---
name: tauri-desktop
description: Keep desktop UI and native backend boundaries clear.
applies_to:
  - tauri
  - desktop
  - frontend
agents:
  - planner
  - coder
  - reviewer
required_evidence:
  - native_boundary
  - ui_boundary
  - desktop_test
---

# Skill: Tauri Desktop

## When to use
Use when editing the Tauri desktop shell or workbench UI.

## Workflow
1. Keep native backend calls behind service boundaries.
2. Keep React flow state separate from native commands.
3. Validate desktop build or record why it is blocked.
4. Preserve local-first behavior.

## Exit criteria
- Native boundary is identified.
- UI boundary is identified.
- Desktop test or blocked reason is attached.
