# Project-Level Agent Preferences

## Instruction Precedence
- System and tool instructions override all other guidance.
- Repository-level `AGENTS.md` or `CLAUDE.md` overrides user-level `~/.claude/CLAUDE.md`.
- This file provides default preferences when project-level guidance is absent.

## Communication
- Respond to the user in Chinese unless the project requires otherwise.
- Keep explanations concise and practical.
- Ask for confirmation only when a decision is high-impact, irreversible, or has non-obvious tradeoffs.

## Execution
- Read existing code, project conventions, and relevant files before making changes.
- Prefer minimal, targeted edits over broad refactors.
- Prefer adapting existing patterns over introducing new structures.
- Do not overwrite or revert user changes unless explicitly asked.

## Workflow Routing
- Use `superpowers:brainstorming` before any creative work, feature design, or behavior modification.
- Use `superpowers:writing-plans` after brainstorming to create an implementation plan.
- Use `superpowers:verification-before-completion` before claiming work is complete or creating a PR.
- Use `superpowers:systematic-debugging` when encountering any bug, test failure, or unexpected behavior.
- Use `superpowers:test-driven-development` when implementing any feature or bugfix, before writing implementation code.
- Use `superpowers:executing-plans` when you have a written implementation plan to execute in a separate session with review checkpoints.
- Use `superpowers:finishing-a-development-branch` when implementation is complete, all tests pass, and you need to decide how to integrate the work.

## Default Tech Context
- Primary stack: Go 1.22+, Bubble Tea TUI, multi-platform CLI tool
- Always prefer project-level `AGENTS.md` / `CLAUDE.md` when available

## Validation
- Run the most relevant tests or checks after changes when feasible.
- If validation is skipped or incomplete, state what was not verified and why.
- Clearly mention assumptions, blockers, and residual risks when relevant.
