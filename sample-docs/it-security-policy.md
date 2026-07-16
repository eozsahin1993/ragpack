# IT and Security Policy

**Solstice Labs — Employee Handbook**
Effective date: January 1, 2026

## Account Security

### Passwords

All company accounts require a password of at least **14 characters**, and password reuse across the last 10 passwords is blocked automatically. Passwords must be rotated every **180 days** for systems that don't support single sign-on.

### Multi-Factor Authentication

Multi-factor authentication (MFA) is **required** for all employees on:

- Company email and calendar
- The HR portal and payroll system
- Source control (GitHub) and cloud infrastructure (AWS, GCP)
- The VPN

Employees who lose access to their MFA device should contact IT immediately at `it-help@solsticelabs.example` for identity verification and reset.

### Single Sign-On

Most internal tools are provisioned through the company SSO provider. New tool requests go through the IT Portal and are typically provisioned within 2 business days.

## Device Policy

### Company-Issued Devices

Company laptops come pre-configured with disk encryption, endpoint monitoring, and automatic security updates. Employees may not disable these tools. Lost or stolen devices must be reported to IT within **1 hour** of discovery so they can be remotely locked.

### Personal Devices (BYOD)

Personal phones may be used to access company email and Slack after enrolling in the Mobile Device Management (MDM) program, which enforces a passcode and remote-wipe capability on the company data partition only. Personal laptops may not be used to access source control, production systems, or customer data.

## Data Handling

### Classification

Company data falls into three tiers:

1. **Public** — marketing materials, published documentation.
2. **Internal** — internal docs, meeting notes, non-sensitive project data.
3. **Restricted** — customer data, financial records, employee PII, source code.

Restricted data may not be stored on personal devices, personal cloud storage, or unapproved SaaS tools. Use only company-approved storage (Google Drive under the company workspace, or the internal data platform).

### Data Retention

Restricted data is retained according to the schedule set by Legal and Security, generally **7 years** for financial records and **3 years** for customer support data after account closure, unless a longer period is required by contract or law.

## Travel Restrictions

Employees traveling to countries on the restricted-travel list (maintained by Security and available on the IT Portal) must use a loaner laptop rather than their primary device, and should contact Security before departure for a travel security briefing.

## Reporting a Security Incident

Suspected security incidents, including phishing emails, lost devices, or suspicious account activity, should be reported immediately to `security@solsticelabs.example` or the `#security-incidents` Slack channel. Do not wait for confirmation that something is actually a problem — report anything suspicious right away.
