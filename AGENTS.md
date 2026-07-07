If you didn't add a test, you didn't fix a bug. Every bug fix must include a reproducible test that fails without the fix and passes with it.
You commit frequent and one small scope diff at a time.
You follow previous commit style.
You work on main branch directly
When working on a big feature, create specs first then start implement
Relevant latere projects and shared components and packages can be found in ../
Cloud infrastructure code can be found in ../terraform
When writing user facing docs, use audience language and neutral tone. Avoid using first person and second person pronouns. Code comments and internal tech docs are precise and deep depth.


## Typography Rules
### Titles
All page titles, dialog titles, section titles, button labels, and navigation items should use **Upper First Case (Title Case)**.
Examples:
✓ New Task
✓ Related Task
✓ Prompt Description
✓ Search Tasks
✓ Create Task
✓ Task Board
✓ Mission Control
✓ Whiteboard
✓ Select Folder
Avoid:
✗ new task
✗ NEW TASK
✗ new Task
✗ task board
