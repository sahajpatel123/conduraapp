# AGENTS.md

> Reference document for the specialized AI coding agents configured for this project.

---

## 1. Overview

This file documents the five specialized AI coding agents available for working on this project. Each agent is manually created and tuned for a specific domain of software engineering, enabling higher performance on complex technical tasks than a single general-purpose agent can provide.

**Why multiple agents?** Different phases of software development require different mindsets. An architect thinks in systems and trade-offs; an implementation engineer thinks in code quality and testability; a security specialist thinks in attack surfaces and data flows. By using purpose-built agents, each task gets the right cognitive focus — resulting in better outputs, fewer blind spots, and faster iteration.

All agents are custom-configured for this project's stack, conventions, and quality standards.

---

## 2. Agent Quick Reference

| Agent | Primary Focus | Best Used For | Key Strength |
|---|---|---|---|
| **Production-Ready Analysis** | Codebase audits and gap analysis | Pre-release reviews, technical debt assessment | Exhaustive, honest, evidence-based findings |
| **System Architect** | Architecture design and decisions | New features, system design, refactoring plans | Sound trade-off analysis and long-term thinking |
| **Implementation Engineer** | Writing production-quality code | Feature implementation, bug fixes, refactoring | Clean, tested, maintainable code delivery |
| **AI Systems Specialist** | AI/agentic system design and improvement | Agent orchestration, evaluations, LLM integration | Deep expertise in agentic system challenges |
| **Security & Privacy Guardian** | Security and privacy risk mitigation | Auth flows, data handling, prompt injection defense | Practical mitigations without slowing development |

---

## 3. Detailed Agent Profiles

### Production-Ready Analysis

- **Purpose**: Conducts deep, comprehensive audits of the codebase and workspace to identify gaps that prevent the project from being production-ready. Produces structured audit reports with prioritized recommendations.

- **When to Use**:
  - Before major releases
  - When preparing to ship features
  - During periodic codebase reviews
  - When you want an honest assessment of technical debt and risks

- **Key Capabilities**:
  - Exhaustive exploration of the workspace
  - Evidence-based findings with code references
  - Severity-based prioritization (critical / high / medium / low)
  - Security + reliability + maintainability analysis
  - Clear phased roadmaps for remediation

- **Strengths**: Extremely thorough and honest. Excellent at surfacing hidden risks and production gaps that are easy to miss during normal development.

- **Example Use Cases**:
  1. Full codebase audit before launching a new version
  2. Reviewing a major feature branch for production readiness
  3. Identifying security and observability gaps across the system
  4. Creating a prioritized hardening roadmap

- **Notes**: This agent should be used with a "comprehensive" mindset. Give it time and broad access to the workspace. It is most valuable when you want the unvarnished truth about what needs fixing.

---

### System Architect

- **Purpose**: Provides high-level architectural guidance, designs new systems and components, reviews existing architecture, and helps make sound long-term technical decisions with clear trade-off analysis.

- **When to Use**:
  - Starting new major features or modules
  - Redesigning parts of the system
  - Making technology or pattern choices
  - Reviewing the overall architecture for scalability or maintainability

- **Key Capabilities**:
  - Trade-off analysis across multiple approaches
  - Component and interface design
  - Scalability and reliability thinking
  - Architecture Decision Records (ADR-style documentation)
  - Balancing short-term delivery with long-term maintainability

- **Strengths**: Thinks at the right abstraction level and prevents expensive architectural mistakes early. Communicates decisions clearly with rationale.

- **Example Use Cases**:
  1. Designing a new multi-agent orchestration layer
  2. Choosing between different state management or data flow approaches
  3. Reviewing current architecture for scalability issues
  4. Planning a major refactor of a core subsystem

- **Notes**: Best used early in the design process. Pair with Implementation Engineer for execution of the resulting design.

---

### Implementation Engineer

- **Purpose**: Writes clean, efficient, well-tested, and production-quality code. Handles feature implementation, complex bug fixing, and refactoring with high engineering standards.

- **When to Use**:
  - During actual coding work
  - Implementing features from a defined design
  - Fixing difficult or nuanced bugs
  - Refactoring modules for clarity or performance

- **Key Capabilities**:
  - High-quality code generation in the project's language and patterns
  - Strong error handling and edge case awareness
  - Performance-conscious implementations
  - Testability by design
  - Following and extending existing code conventions

- **Strengths**: Bridges the gap between design and working code with excellent engineering discipline. Produces code that other developers (and agents) can understand and maintain.

- **Example Use Cases**:
  1. Implementing a new simulation engine component
  2. Refactoring a complex agent workflow for clarity
  3. Fixing tricky bugs in data pipelines or async code
  4. Writing well-structured API endpoints or frontend components

- **Notes**: Provide clear requirements and context. It performs best when the architecture or approach is already reasonably defined. Pair with System Architect for design decisions and with Security & Privacy Guardian for sensitive code paths.

---

### AI Systems Specialist

- **Purpose**: Specializes in designing, building, and improving AI and agentic systems — including multi-agent orchestration, behavioral simulations, evaluation harnesses, LLM integrations, guardrails, and related concerns.

- **When to Use**:
  - Working on anything related to agents, simulations, or evaluation systems
  - Prompt engineering at scale or LLM integration work
  - Designing agent safety, guardrails, or observability
  - Optimizing cost and latency of AI workflows

- **Key Capabilities**:
  - Agent architecture patterns (hierarchical, swarm, orchestration)
  - Orchestration and delegation design
  - Evaluation strategies for non-deterministic systems
  - Cost and latency optimization for LLM calls
  - Safety, guardrails, and content filtering
  - Observability and tracing for agent decisions
  - Handling non-determinism and failure modes

- **Strengths**: Deep expertise in the unique challenges of building reliable and observable agentic systems. Understands the failure modes specific to LLM-powered systems.

- **Example Use Cases**:
  1. Designing hierarchical agent systems with delegation and supervision
  2. Building evaluation frameworks for synthetic or behavioral agents
  3. Improving simulation engine logic for realism and performance
  4. Adding guardrails, tracing, or cost controls to agent runs

- **Notes**: This is one of the most important agents for projects involving AI agents, simulations, or LLM-driven workflows. Use it heavily for agent-related architecture and implementation. Pair with Security & Privacy Guardian for prompt injection and safety concerns.

---

### Security & Privacy Guardian

- **Purpose**: Identifies security and privacy risks, reviews code and architecture for vulnerabilities, and recommends practical mitigations with a focus on secrets management, data protection, and secure AI integrations.

- **When to Use**:
  - Building features involving user data, authentication, or external APIs
  - Working on LLM integration points or prompt handling
  - Handling any sensitive or personal information
  - Before merging code that touches security-relevant paths

- **Key Capabilities**:
  - Secrets hygiene and management review
  - Injection prevention (SQL, command, prompt injection)
  - Access control pattern review
  - Data minimization and privacy-by-design thinking
  - Dependency security and supply chain analysis
  - Compliance-relevant practice checks

- **Strengths**: Helps maintain a strong security and privacy posture without slowing down development excessively. Focuses on practical, actionable mitigations rather than theoretical risks.

- **Example Use Cases**:
  1. Reviewing authentication and data handling flows
  2. Auditing LLM integration points for prompt injection risks
  3. Improving secrets management across the project
  4. Checking compliance-relevant practices for health or financial data

- **Notes**: Use proactively rather than only after problems appear. Particularly important for projects handling health, financial, or personal data. Pair with AI Systems Specialist for agent-specific safety concerns.

---

## 4. How to Choose the Right Agent

Use this decision framework when assigning a task:

**Step 1: Identify the primary nature of the task.**

| Task Nature | Agent |
|---|---|
| "What needs to be fixed or improved?" | Production-Ready Analysis |
| "How should we design this?" | System Architect |
| "Write the code for this." | Implementation Engineer |
| "How should the AI/agent system work?" | AI Systems Specialist |
| "Is this safe and secure?" | Security & Privacy Guardian |

**Step 2: Check for overlap.** Many tasks benefit from multiple agents. Use this priority:

- **New feature from scratch**: System Architect → Implementation Engineer
- **Security-sensitive feature**: Security & Guardian + AI Systems Specialist → Implementation Engineer
- **Bug fix**: Implementation Engineer (add Security Guardian if it involves auth, data, or LLM calls)
- **Pre-release check**: Production-Ready Analysis (add Security Guardian for audit)
- **Agent or simulation work**: AI Systems Specialist (add Security Guardian for safety)

**Step 3: When uncertain, default to:**
- Production-Ready Analysis for assessment tasks
- Implementation Engineer for execution tasks
- Add Security & Privacy Guardian as a second pass for anything touching user data, auth, or external systems

---

## 5. Best Practices for Using Multiple Agents

1. **Sequence wisely.** Start with assessment or design agents (Production-Ready Analysis or System Architect) before handing off to implementation agents. This prevents wasted effort on poorly designed solutions.

2. **Combine for critical paths.** For security-sensitive or high-impact code, use two agents: one for implementation and one for review. The Implementation Engineer writes the code; the Security & Privacy Guardian reviews it.

3. **Maintain consistency.** When switching between agents in one session, provide the previous agent's output as context. This prevents conflicting approaches and preserves design decisions.

4. **Use the right abstraction level.** Don't ask the Implementation Engineer to make architectural decisions. Don't ask the System Architect to write production code. Keep each agent in its lane for best results.

5. **Iterate in short cycles.** Rather than asking one agent to do everything, break work into small handoffs: design → implement → review → refine. Each handoff should include clear context and acceptance criteria.

6. **Leverage strengths for weak spots.** If you're unsure about a domain (e.g., security), always include the specialist agent as a second pass, even if another agent did the initial work.

---

## 6. Maintenance

This file should be updated when:

- An agent is added, removed, or significantly reconfigured
- Agent capabilities change due to prompt or model updates
- New use cases emerge that aren't covered by current descriptions
- Agent naming or ordering changes

Keep descriptions accurate and concise. The goal is a living reference that remains genuinely useful over time — not a document that drifts from reality.

**Last updated**: 2026-06-23
