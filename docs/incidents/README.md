# Incident Reports (RCA)

This directory contains the **Root Cause Analysis (RCA)** and post-mortem reports for service disruptions, bugs, or security incidents within the Uptime Monitor.

---

## Incident Log

No incidents recorded yet.

---

## Process & Standards

Incident documentation prevents recurrence and builds system resilience.

### When to write an RCA (The Rule of Three)

Formal RCAs are required only if **one or more** of these conditions are met:

1. **Utility Loss**: Failure to fulfill primary purpose, such as uptime checks stopping, dashboard data becoming unavailable, or alerts failing.
2. **Data Integrity**: Permanent loss, corruption, or unauthorized exposure of monitoring data.
3. **Regression (The "Zombie Bug")**: The failure has occurred previously. Identification of the gap in the previous fix is required.

Minor configuration drift or noisy logs that do not impact system health should be handled via standard Git commit documentation rather than an RCA.

### Severity Levels

| Level | Meaning |
| :--- | :--- |
| **🔴 High** | Service down, data loss, or security breach. |
| **🟡 Medium** | Partial degradation, performance issues, or feature malfunction. |
| **🔵 Low** | Minor bugs, cosmetic issues, or non-critical failures. |

### Status

| Status | Meaning |
| :--- | :--- |
| **🚧 Investigating** | Identifying the root cause. |
| **🩹 Mitigated** | Temporary fix applied, service restored. |
| **✅ Resolved** | Root cause identified and permanent fix implemented. |

### RCA Template

To document a new incident, create a new file named `XXX-descriptive-title.md`.

```markdown
# RCA [XXX]: [Descriptive Title]

- **Status:** 🚧 Investigating | 🩹 Mitigated | ✅ Resolved
- **Date:** YYYY-MM-DD
- **Severity:** 🔴 High | 🟡 Medium | 🔵 Low
- **Author:** Victoria Cheng

## Summary

A brief overview of what happened, the impact, and the duration.

## Timeline

- **YYYY-MM-DD HH:MM:** Incident detected.
- **YYYY-MM-DD HH:MM:** Investigation started.
- **YYYY-MM-DD HH:MM:** Mitigation applied.
- **YYYY-MM-DD HH:MM:** Root cause identified.
- **YYYY-MM-DD HH:MM:** Permanent fix deployed.

## Root Cause Analysis

Detailed explanation of why the incident happened.

## Lessons Learned (Optional)

What went well? What went wrong? What reduced the impact?

## Action Items

- [ ] **Fix:** Immediate technical resolution.
- [ ] **Prevention:** Changes to prevent recurrence, such as monitoring or tests.
- [ ] **Process:** Changes to workflows or documentation.

## Verification

- [ ] **Manual Check:**
- [ ] **Automated Tests:**
```
