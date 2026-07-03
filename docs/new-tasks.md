# Feature Update — AI Quick Capture

See [docs/imgs/img.png](docs/imgs/img.png) mockup step flow. 

Replace the current "Create Task" modal with an AI-first Quick Capture workflow.

The current modal is a traditional CRUD form:

- Title
- Description
- Priority
- Assigned Agent

This is not suitable for an AI Development Workbench.

Creating a task should feel like talking to an AI instead of filling a Jira form.

---

# Goal

Users should be able to paste anything.

Examples:

- Plain text
- Screenshot
- UI mockup
- Figma link
- GitHub issue
- Jira URL
- Markdown
- Requirements document
- Multiple images
- Mixed text + images

The AI should automatically convert the input into a structured engineering task.

---

# UX Philosophy

Users should NOT think about:

- Title
- Labels
- Priority
- Acceptance Criteria
- User Story
- Technical Notes

The planner agent should infer these automatically.

Users only review and edit before creating the task.

---

# New Flow

Current

New Task

↓

Form

↓

Create

↓

Board

New

New Task

↓

Quick Capture

↓

AI Structure

↓

Review & Edit

↓

Create Task

↓

Planning Starts

---

# Wizard

Use a 4-step wizard.

## Step 1

Quick Capture

AI Powered

Large textarea.

Placeholder

Paste anything...

Examples:

• User requirements
• Screenshots
• Figma links
• GitHub Issues
• Markdown
• Product specs

Support

✓ Paste text

✓ Paste image

✓ Drag & Drop

✓ Upload image

✓ Paste URL

Supported shortcuts

⌘V

Ctrl+V

Drag files

Example chips

Build shopping cart

Fix login bug

Dark mode

Search feature

Payment integration

Bottom button

AI Structure →

---

## Step 2

AI Structure

Planner analyzes the input.

Automatically generate

Task Title

Description

Feature

Epic

Priority

Labels

Acceptance Criteria

Technical Notes

Likely Files

Estimated Scope

Potential Risks

Display confidence for each field.

Example

Title

Implement Shopping Cart

Priority

P1

Confidence

95%

Acceptance Criteria

✓ Add product

✓ Remove product

✓ Update quantity

✓ Checkout

Editable.

Users can modify everything.

---

## Step 3

Review & Edit

Multiple sections.

Task Details

Attachments

AI Plan Preview

Task Details

Title

Description

Priority

Labels

Estimate

Owner

Attachments

Image previews

Figma

GitHub

Markdown

AI Plan Preview

Planner Summary

Initial execution plan

Likely files

Open questions

Missing information

If AI lacks information

Show

Missing Information

• Payment provider?

• Mobile support?

• Guest checkout?

Planner can ask questions later.

---

## Step 4

Create

Summary

Everything ready.

Button

Create Task

After creation

Immediately open the task.

Planning starts automatically.

---

# Planning Flow

Task Created

↓

Planning Step

↓

Launch configured Planning Agent

↓

Open native terminal

↓

Planner asks questions

↓

User replies

↓

Execution Plan

↓

Waiting Approval

No coding starts automatically.

---

# Attachments

Support

Images

PNG

JPG

WebP

PDF

Markdown

Text

URLs

Figma

GitHub

Jira

Display previews.

Allow remove.

Allow reorder.

---

# AI Extraction

Planner should extract

Title

Description

User Story

Acceptance Criteria

Priority

Labels

Technical Notes

Likely Files

Risks

Dependencies

Suggested Workflow

Confidence

Never require all fields manually.

---

# Empty Examples

Provide example prompts.

Example 1

Implement shopping cart.

Need same UI as attached screenshot.

Guest checkout.

Responsive.

Example 2

Fix login bug.

Attached console log.

OAuth redirects incorrectly.

Example 3

Implement design from Figma.

Support desktop and mobile.

---

# Smart Features

Support

Paste Screenshot

↓

AI OCR

↓

Extract requirements

Support

Paste Figma

↓

Extract URL

↓

Attach to task

Support

Paste GitHub Issue

↓

Link issue

↓

Import summary

---

# UI

Dark mode.

TailwindCSS.

Rounded cards.

Purple accent.

Use shadcn/ui primitives.

Use lucide-react icons.

Icons

Sparkles

ClipboardPlus

ImagePlus

Link

Bot

Check

FileText

ArrowRight

---

# Components

QuickCaptureFlow

QuickCaptureEditor

AttachmentDropzone

ImagePreviewGrid

AIExtractionCard

ConfidenceBadge

TaskReviewFlow

PlanPreviewCard

MissingInformationCard

CreateTaskWizard

---

# Flow Component Pattern

flows/

quick-capture-flow/

QuickCaptureFlow.tsx

QuickCaptureEditor.tsx

AIExtractionFlow.tsx

ReviewTaskFlow.tsx

CreateTaskFlow.tsx

useQuickCaptureFlow.ts

features/

tasks/

shared/

ui/

---

# Implementation Scope

Implement frontend only.

Use mocked AI extraction.

Do not call real LLM.

Do not start planner yet.

Return structured mock data.

The backend integration will be implemented in a later phase.

---

# Acceptance Criteria

✓ Traditional CRUD form is removed.

✓ New Task opens Quick Capture wizard.

✓ Users can paste text, images and links.

✓ AI Structure screen generates editable task fields.

✓ Review screen supports editing.

✓ Attachments display previews.

✓ Flow feels like an AI assistant instead of Jira.

✓ Responsive.

✓ TailwindCSS.

✓ Flow Component Design Pattern.