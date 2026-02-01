# Linear Quick Reference - Living Lands Discord Bot

**Team:** Living Lands Discord Bot (LLB)  
**Board:** https://linear.app/moshpitcodes/team/LLB/active

---

## 📊 Custom Views (Quick Access)

### 0.1.0 MVP Tracker
**Description:** MVP scope for initial launch (linking + welcome + /guide + /ask skeleton)

**Suggested Labels:**
- `v0.1.0`
- `mvp`
- `infra`
- `discord`
- `llm`
- `rag`

### Infra
**Description:** Docker, config, migrations, observability

### Critical & Urgent
**Description:** Priority 1-2 issues blocking the bot

### Testing & Reliability
**Description:** Unit/integration tests, reliability, rate limits

---

## 🏷️ Filtering by Labels

### v0.1.0 MVP Issues
**Filter:** `label:v0.1.0`  
**URL:** https://linear.app/moshpitcodes/team/LLB/active?filter=label%3Av0.1.0

### Discord Integration
**Filter:** `label:discord`

### LLM
**Filter:** `label:llm`

### RAG
**Filter:** `label:rag`

### Infra
**Filter:** `label:infra`

---

## 🔄 Cycles

Define cycles per release once the LLB team has cycles configured.

---

## 📊 Priority Levels

Linear defaults:
- Priority 1: Urgent
- Priority 2: High
- Priority 3: Medium
- Priority 4: Low

---

## 🎯 Quick Actions

### Access Custom Views (FASTEST METHOD)
1. Go to Linear: https://linear.app/moshpitcodes/team/LLB/active
2. Click **"Views"** in left sidebar
3. Select view:
   - **v1.3.0 Release Tracker** - All v1.3.0 issues (8 issues)
   - **Announcer Module** - Announcer work only (4 issues)
   - **Critical & Urgent** - High-priority work (9 issues)
   - **Testing & Performance** - Testing/perf issues (4 issues)

### Filter v0.1.0 Issues
Use label filter:
1. Go to: https://linear.app/moshpitcodes/team/LLB/active
2. Click filter icon → Select label: `v0.1.0`

### View Completed Work (Retrospective)
**Filter:** `status:done`  
**Shows:** 9 completed issues (8 retrospective + 1 previous)

---

## 🏗️ Issue Naming Conventions

**Format:** `[Area] Feature/Task Description`

**Examples:**
- `[Infra] Docker compose + healthchecks`
- `[Discord] Implement /link command`
- `[API] Verify endpoint for Hytale`
- `[LLM] Ollama generate client`
- `[RAG] Index docs into ChromaDB`

---

## 📋 Labels Reference

| Label | Description | Usage |
|-------|-------------|-------|
| `v0.1.0` | MVP release issues | Version tracking |
| `mvp` | MVP scope | Scoping |
| `infra` | Docker/config/migrations | Area |
| `discord` | Discord commands/events | Area |
| `api` | Fiber endpoints | Area |
| `llm` | Ollama integration | Area |
| `rag` | Chroma indexing/querying | Area |
| `documentation` | Docs work | Issue type |

---

## 🔍 Search Tips

### Find Issues by Module
- `label:announcer`
- `label:professions`
- Future: `label:economy`, `label:claims`, `label:groups`

### Find Issues by Release
- `label:v1.3.0` (current release)
- Future: `label:v1.4.0`, `label:v1.5.0`, `label:v1.6.0`

### Combine Filters
- `label:v1.3.0 priority:urgent` → Critical blockers for v1.3.0
- `label:announcer status:backlog` → Announcer work not started
- `label:testing estimate:>0` → Testing issues with estimates

---

## 📅 Weekly Planning

### Week 3 (Current)
1. Manually assign v1.3.0 issues to cycle (5 min)
2. Link v1.3.0 dependencies (30 min)
3. Start multi-player testing planning (MPC-87)

### Week 4-6
1. Execute multi-player stress testing (MPC-87)
2. Begin parallel work (MPC-84-86, Announcer)
3. Add estimates to v1.4.0+ issues

---

## 🔗 External Links

 - **Linear Board:** https://linear.app/moshpitcodes/team/LLB/active

---

---

## ✨ Custom Views Summary

| View Name | Color | Purpose | Issues |
|-----------|-------|---------|--------|
| **v1.3.0 Release Tracker** | Green #0E8A16 | Current release scope | 8 |
| **Announcer Module** | Orange #F2994A | Announcer feature work | 4 |
| **Critical & Urgent** | Red #EB5757 | Priority 1-2 issues | ~10 |
| **Testing & Performance** | Yellow #F2C94C | Testing/perf work | 4 |

All views are **shared** with the team for collaborative planning.

**How to Access:** Linear UI → Left Sidebar → "Views" → Select view name

---

**Last Updated:** 2026-01-31 (✨ Custom views added!)  
**Current Release:** v1.2.3 (Production Ready)  
**Next Release:** v1.3.0 - Testing & Performance (Feb 1 - Mar 15, 2026)
