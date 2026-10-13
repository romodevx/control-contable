# Security policy

Control Contable handles Mexican taxpayer data (RFC, CURP, income and expense records), so we
take reports seriously and answer them privately.

## Reporting a vulnerability

**Do not open a public issue, pull request or discussion for a security problem.**

Report it through GitHub's private vulnerability reporting:
**Security → Report a vulnerability** on this repository
(<https://github.com/romodevx/control-contable/security/advisories/new>).

Please include:

- what an attacker can do and what they need first (an account, a second account, network position);
- the affected component (`backend/`, `mobile/`, `infra/`, CI) and the commit or version;
- steps to reproduce, ideally with a minimal request or test case;
- any logs or screenshots, with real personal data removed.

What to expect:

| Step | Target |
|---|---|
| Acknowledgement | within 3 business days |
| First assessment (severity, affected versions) | within 10 business days |
| Fix or mitigation for a confirmed high or critical issue | within 30 days |

We credit reporters in the advisory unless you ask us not to.

## Scope

In scope: this repository's code and configuration, including CI workflows and the local
stack under `infra/local`.

Out of scope:

- the local-only placeholder passwords in `infra/local/.env.example` and the throwaway
  credentials in tests (they guard disposable containers bound to loopback);
- findings that need a compromised developer machine or a malicious maintainer;
- denial of service through traffic volume alone;
- third-party services (Google Cloud, RevenueCat, app stores): report those to the vendor.

## Safe harbour

We will not pursue legal action for research that stays within this policy: test only against
your own local deployment, never access or modify other people's data, and give us reasonable
time to fix the issue before disclosure.

## Supported versions

The project is pre-release. Only the `dev` branch (and `main` once releases begin) receives
security fixes.
