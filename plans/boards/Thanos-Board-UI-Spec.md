# Thanos Board UI Spec

## Goal

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

### Prompt / Description

Large textarea or markdown editor.

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

Required: - Title - Prompt / Description

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
