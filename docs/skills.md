# Skills

A **skill** is a reusable, self-contained instruction set that an [agent](./agents.md) can activate to handle a specific type of task. Skills follow the [Agent Skills open standard](https://agentskills.io) and are defined in `SKILL.md` files.

## Anatomy of a skill

A `SKILL.md` file consists of YAML frontmatter and a markdown body:

```markdown
---
name: my-skill
description: A brief description of what this skill does.
---

Detailed instructions for the agent when this skill is activated.
```

| Frontmatter field | Constraint | Purpose |
|-------------------|-----------|---------|
| `name` | Required, alphanumeric + hyphens, ≤ 64 chars | Unique identifier. |
| `description` | Required, ≤ 1 024 chars | Shown to the agent to decide whether to activate. |
| `license` | Optional | License identifier. |
| `compatibility` | Optional, ≤ 500 chars | Notes on which agents or environments are supported. |

The markdown body below the frontmatter contains the full instructions the [agent](./agents.md) receives when the skill is activated. Skills may reference companion files in `scripts/`, `references/`, or `assets/` subdirectories relative to the `SKILL.md`.

Implementation: `internal/skills/skills.go`.

## Discovery

Skills are discovered automatically at startup from multiple directories:

| Scope | Paths |
|-------|-------|
| Global | `~/.config/crush/skills/`, `~/.config/agents/skills/` |
| Global (Windows) | `%LOCALAPPDATA%/crush/skills/`, `%APPDATA%/agents/skills/` |
| Project | `.crush/skills/`, `.agents/skills/`, `.claude/skills/`, `.cursor/skills/` |

Additional directories can be added via `options.skills_paths` in [configuration](./configuration.md). Discovery uses a concurrent file walker (`charlievieth/fastwalk`) that follows symlinks.

## Built-in skills

Crush ships embedded skills compiled into the binary. Built-in skills use virtual paths prefixed with `crush://skills/`. When a user-defined skill shares a name with a built-in, the user skill takes precedence.

## Activation

Skills are not automatically executed. Instead, their names and descriptions are included in the [system prompt](./system-prompts.md) as XML blocks (via `ToPromptXML()`). When the [agent](./agents.md) decides a skill is relevant, it reads the skill's `SKILL.md` using the [view tool](./tools.md) to load the full instructions.

## Filtering

Individual skills can be disabled via `options.disabled_skills` in [configuration](./configuration.md). Duplicate skill names are resolved by `Deduplicate()` — the last-discovered instance wins (user skills override built-ins).

## Related

- [System Prompts](./system-prompts.md) – skill summaries are injected into the prompt.
- [Tools](./tools.md) – the view tool is used to read skill instructions.
- [Context Files](./context-files.md) – a complementary mechanism for project-wide instructions.
- [Configuration](./configuration.md) – skill paths and disabled skills are configured here.
- [Agents](./agents.md) – decide when to activate a skill.
