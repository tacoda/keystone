# Prompt Injection

A coding agent's instructions, capabilities, and data flow through the same channel: text. The agent has no native ability to distinguish "instructions from the user" from "data the agent is reading" — both arrive as tokens. Prompt injection is the class of attack that exploits this: an attacker arranges for the agent to read content that *contains instructions*, expecting the agent to confuse them with directives from its principals.

> **Rules extracted:** [`guides/principles/prompt-injection.md`](../../guides/principles/prompt-injection.md).

## What it asks of you

- Treat the trust boundary as *between channels*, not within them. The user's typed message is high-trust; everything the agent *reads on the user's behalf* (a file, a webpage, a tool response) is lower-trust.
- Reject instruction-shaped content from low-trust sources even when it appears polite, plausible, or aligned with the task. Politeness is not authorization.
- Reduce blast radius by capability. An agent reading an unverified URL does not also need to be holding write access to the filesystem at that moment.

## Why it holds

Greshake et al. (2023) named *indirect prompt injection* — content placed in a document, webpage, or tool response that an LLM later reads and follows. The attack works because the agent has no schema for "instructions from the operator" vs. "data the operator pointed me at." Simon Willison's *prompt injection* writing (2022–) extended this to the day-to-day failure modes: agents leaking system prompts via crafted user input, agents executing tool calls on instructions hidden inside HTML pages, agents exfiltrating context by following links in email bodies.

OWASP's *LLM Top 10* (2023, ongoing) lists prompt injection as the #1 risk for LLM-powered applications, specifically because mitigations are partial and the impact compounds with the agent's tool authority. The 2024 wave of MCP-integrated agents made this concrete: any tool that returns external content (Notion, Linear, Slack, web fetch) is a potential injection surface.

The defense literature is asymmetric. Detection is hard (the attack uses natural language, not a fixed signature); pure prevention is harder (the agent must read something to be useful). The practical mitigations are structural: small capability surfaces during read-untrusted operations, explicit instruction-vs-data discipline at the agent level, human-in-the-loop confirmation for any action triggered by read content.

## Anti-patterns

- "But the tracker card *said* to do it" — the card is data; the user is the operator.
- An agent that auto-runs commands found in a README it just installed.
- An agent that follows links in PR comments without surfacing them.
- A system that loads tool responses directly into the agent's instruction context without a trust boundary.
- An MCP server that returns `system_prompt`-shaped responses as plain content.
- An agent that "tests" instructions from a webpage to "see what they do."

## References

- Greshake, K., Abdelnabi, S., Mishra, S., Endres, C., Holz, T., & Fritz, M. (2023). *Not what you've signed up for: Compromising Real-World LLM-Integrated Applications with Indirect Prompt Injection*.
- Willison, S. (2022–). *Prompt injection* article series. [simonwillison.net/tags/prompt-injection/](https://simonwillison.net/tags/prompt-injection/).
- OWASP. *LLM Top 10* — LLM01: Prompt Injection. [genai.owasp.org](https://genai.owasp.org/).
- Perez, F., & Ribeiro, I. (2022). *Ignore Previous Prompt: Attack Techniques For Language Models*.
- Anthropic. [*Constitutional AI*](https://www.anthropic.com/research/constitutional-ai) and tool-use safety writing — boundaries between operator and external content.

---

Forward link: [`guides/principles/prompt-injection.md`](../../guides/principles/prompt-injection.md). See also: [`security-threats.md`](security-threats.md), [`secrets-management.md`](secrets-management.md).
