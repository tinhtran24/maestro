# Thanos Board UI Spec

## Goal
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

Improve the Board screen into a focused AI task execution board with a
lightweight New Task sidebar.

## Main Changes

-   Remove Task Board header block.
-   Remove inline task title input.
-   Remove inline prompt input.
-   Remove board summary section.
-   Keep only board controls, columns, and task cards.
-   New task creation opens in a right sidebar drawer.

## Layout

``` txt
Sidebar | Top Bar
        | Search / Filters / View Controls
        | Kanban Board Columns       | New Task Drawer
```

## Board Controls

Top control row should include:

-   Search tasks
-   Status filter
-   Label filter
-   Priority filter
-   Group by Status
-   View toggle

Do not add large page titles or summary cards.

## Add Task Behavior

When user clicks:

``` txt
+ Add task
```

inside any column:

-   Open the right New Task sidebar.
-   Preselect the clicked column status internally.
-   Keep the board visible.
-   Do not open modal.
-   Do not navigate away.

## New Task Sidebar

Width: **420--480px**

Position: - Right-side fixed drawer - Full height below top bar

Header:

``` txt
New Task        X
```

Fields:

``` txt
Title *
Prompt / Description *
Related task
```

Remove: - Due date - Assignee - Estimate - Complex settings - Tabs -
Extra metadata

## New Task Form

### Title

Single-line input.

Placeholder:

> Enter a clear, concise title...

###  User Requirements / Description / Prompt

Markdown editor.
Image Input

Placeholder:

> Describe what needs to be done and how we'll know it's complete...

### Related Task

Searchable dropdown.

Placeholder:

> Search tasks...

Helper text:

> Link an existing task that this is related to.

## Footer Actions

Sticky footer:

``` txt
Cancel                Create task
```

## Validation

Required: - Title - User Requirements / Description / Prompt

## Keyboard Shortcuts

``` txt
N           Open New Task
Esc         Close drawer
Ctrl+Enter  Create task
/           Focus search
```

## Empty State

``` txt
No tasks here

+ Add task
```

## Visual Style

-   Dark theme
-   Purple accent
-   Rounded cards
-   Compact spacing
-   Linear / Cursor inspired
-   No Jira-style clutter

## Acceptance Criteria

-   Clicking **+ Add task** opens the right sidebar.
-   Sidebar only contains **Title**, **Prompt / Description**, and
    **Related task**.
-   Board header is removed.
-   Board summary is removed.
-   Board remains visible while creating a task.
-   Cancel closes the drawer.
-   Create Task inserts the task into the selected column.
