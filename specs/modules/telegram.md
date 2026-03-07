# Watch

1. start with the `/watch` command.
2. each configured interval (300s), the bot polls Gmail for new emails.
3. when a new email arrives, the bot understands which parser to use:
  4. if sender is known, use the first parser with the matching sender address;
  5. if sender is unknown, pass the email through all prompts in order.
4. When the required sets of prompts are identified, the bot sends the email content to Claude SDK for summarization with prompt specified.
5. The summarization results must have the following fields:
  5.1. result: 'matched' | 'not matched'
  5.2. title: email title
  5.3. content: summary of the email content per prompt
6. If the result is 'matched', there's no need to use further prompts and the bot sends the summary to the configured chat.
7. If the result is 'not matched', the bot continues to the next prompt and repeats the process. If no prompt matches, the email is ignored.

> Example:
  - New job feedback email arrives in Gmail inbox
  - Telegram bot reads the email content and sends it to Claude SDK for summarization based on pre-defined prompt: 'please summarize the recruiter email with the application review/interview results. Answer with a concise summary: green icon if passed, red icon if failed. and the next steps or reasons for the decision.'
  - Claude SDK understands the following:
    - link to the email
    - email content in json format with fields that are set in the prompt configuration
  - Telegram bot forwards parsed summary to configured chat to render it as a well-formatted message.
