---
description: Guide for using the submodule.go DI framework — registration, testing, mocking, scopes, and anti-patterns
allowed-tools: Glob, Grep, Read, Edit, Write, Bash
---

Use the `submodule-go` skill to guide work with the submodule.go dependency injection framework.

**What it does:**
1. Ensures correct DI registration patterns (Make, Resolve, Value, Group, MakeModifiable)
2. Enforces proper testing with isolated scopes and ResolveToWith mocking
3. Prevents anti-patterns (calling Resolve inside factories, re-implementing singletons)
4. Guides middleware usage for lifecycle management

**Arguments:**
- No args: Apply the skill to the current task
- `review`: Audit existing code for submodule.go anti-patterns
- `test`: Help write tests with proper scope isolation and mocking
