---
name: good-skill
description: Builds weekly client finance reports and raises invoices. Use when the user asks for a client report, an invoice, or a refund.
metadata:
  tested-models: [haiku, sonnet, opus]
---

# Client reports

Read [reference/finance.md](reference/finance.md) for revenue questions and
[reference/sales.md](reference/sales.md) for pipeline questions.

Copy this checklist into your reply and tick it off:

- [ ] Step 1: Gather the data
- [ ] Step 2: Draft the report
- [ ] Step 3: Validate the report
- [ ] Step 4: Raise the invoice

## Step 1: Gather the data
Pull the numbers the user asked about.

## Step 2: Draft the report
Write it in the client's voice.

## Step 3: Validate the report
Run `python scripts/validate.py report.md`. Fix every issue it reports and
re-run until it passes.

## Step 4: Raise the invoice
Install the dependency first: `pip install requests`.
Run exactly `python scripts/send_invoice.py --client <id>`; do not add flags.
