## Summary

<!-- What changed and why? Keep this to one concern. Link the issue when applicable. -->

## Compatibility and documentation

- [ ] This preserves backward compatibility, or the linked issue explicitly authorizes the breaking change.
- [ ] Documentation was updated, or no API, CLI, setup, or workflow documentation changed.

## Validation

<!-- List commands run and their results. State why a normally relevant check did not apply. -->

- [ ] Relevant tests were added or updated for behavior changes.
- [ ] Relevant formatting, linting, type checking, build, and test commands pass.

## Delivery details

- Branch: `<approved-prefix>/<short-kebab-case-summary>`
- Conventional Commit: `<type>(<optional-scope>): <imperative summary>`
- Suggested PR title: `<type>(<optional-scope>): <imperative summary>`

## Final review

- [ ] This PR addresses one focused concern and contains no unrelated refactoring.
- [ ] `git diff --check` passes and the final status contains no secrets, local state, temporary files, or unwanted generated/build artifacts.
- [ ] Required generated artifacts were regenerated from source; no generated files were hand-edited.
