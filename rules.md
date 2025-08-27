# Rules YAML File Specification for AI Agents

This document explains the structure and content of the `rules.yaml` file, which is used by the Document Scanning Rules Engine to define patterns for identifying sensitive information or specific data within documents. As an AI agent, you should use this specification when creating or modifying `rules.yaml` files.

## Purpose

The `rules.yaml` file contains a collection of rules, each designed to detect a specific pattern within text. These rules are applied to documents during the scanning process, and any matches are reported as "findings."

## Structure

The `rules.yaml` file must be a YAML dictionary with a single top-level key: `rules`. The value associated with this key is a list of rule objects. Each rule object must have the following keys:

-   `id` (string, **required**): A unique identifier for the rule. This should be a short, descriptive string (e.g., `credit-card-number`, `email-address`, `ssn`).
-   `pattern` (string, **required**): A regular expression (regex) pattern that the engine will use to search for matches within the document text. Ensure the regex is valid and correctly escapes any special characters.
-   `severity` (string, **required**): The severity level of the finding if the rule matches. Common values include `high`, `medium`, `low`, or `informational`.
-   `description` (string, **optional**): A human-readable description of what the rule detects. This is useful for understanding the purpose of the rule and for generating reports.

## Examples

Here are some examples of how rules can be defined in `rules.yaml`:

```yaml
rules:
  - id: credit-card-number
    pattern: "\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|6(?:011|5[0-9]{2})[0-9]{12}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|(?:2131|1800|35\d{3})\d{11})\b"
    severity: high
    description: Detects common credit card numbers (Visa, MasterCard, American Express, Discover, Diners Club, JCB).

  - id: email-address
    pattern: "\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b"
    severity: medium
    description: Identifies standard email address formats.

  - id: social-security-number
    pattern: "\b\d{3}-\d{2}-\d{4}\b"
    severity: high
    description: Detects U.S. Social Security Numbers in XXX-XX-XXXX format.

  - id: api-key-placeholder
    pattern: "API_KEY_PLACEHOLDER"
    severity: informational
    description: Identifies placeholder text for API keys.
```

## Guidelines for AI Agents

When creating or modifying `rules.yaml`, consider the following:

1.  **Uniqueness of `id`**: Ensure that each `id` is unique across all rules in the file.
2.  **Regex Validity**: Always use valid regular expressions. Test your regex patterns thoroughly to avoid errors.
3.  **Specificity vs. Generality**: Balance the specificity of your regex patterns. Too specific, and you might miss variations; too general, and you might get too many false positives.
4.  **Severity Assignment**: Assign severity levels based on the potential impact of the detected information.
5.  **Clear Descriptions**: Provide clear and concise descriptions for each rule, explaining what it aims to detect.
6.  **File Location**: The `rules.yaml` file is typically expected in the root directory of the project or a designated configuration directory.
7.  **Validation**: After generating or modifying the `rules.yaml` file, it's crucial to validate its YAML syntax and test the regex patterns against sample data to ensure they function as expected.

Example:
rules.yaml

- id: profanity-1
  pattern: "badword"
  severity: high
  description: "Detects common profanity"
- id: sensitive-phrase-1
  pattern: "confidential information"
  severity: medium
  description: "Detects sensitive phrases"

8.  **Backslash Escaping in Regex Patterns**: When defining `pattern` fields in YAML, especially for regular expression escape sequences like `\b` (word boundary), `\s` (whitespace), `\d` (digit), etc., you must use *double backslashes* (`\\`). This is because YAML parsers interpret a single backslash as an escape character. For example, `\b` should be written as `\\b`, and `\s` as `\\s`. Failure to do so will result in "unknown escape sequence" errors or incorrect pattern matching when the regex is compiled by the application. This is particularly important when LLMs generate these patterns, as they might not automatically handle YAML-specific escaping.