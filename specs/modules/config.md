# Start
1) User registers in the bot and sends /start. It starts the installation guide telling how to set up GCP OAuth token for email access. `/register`, `/watch` command is available.

# Register
this command is used to link new Gmail account for monitoring.
1) User sends `/register` command. Command shows the installation guide.
2) The bot validates the token is working and saves the GCP OAuth token for email access.
3) The 'configure' button is available below using [Special Keyboard](https://core.telegram.org/bots/features#keyboards) as well as `/configure` command.

# Configure
this command is used to configure filters and prompts for summarization. It's not available if `register` command result is not successful.
1) User sends `/configure` command. Command shows the current prompt configurations (empty after registration).
2) near each prompt, there is a list with following fields:
- sender filter: the specific sender email (optional)
- prompt: the summarization prompt. Will read all emails if sender filter is not specified.
- `edit` button to go to prompt editing page (`/addprompt` command with existing prompt ID and setup fields)
- `remove` button to delete the prompt.
3) There is an `Add new` button at the bottom using [Special Keyboard](https://core.telegram.org/bots/features#keyboards). It's also available as `/addprompt` command.

# AddPrompt
this command is used to add a new prompt for summarization. It's not available if `register` command result is not successful.
1) User sends `/addprompt` command. Command shows the prompt editing page.
2) Two input fields are shown: sender filter (optional) and prompt text and the 'save' button below with labels explaining what each field does.
2) If an existing prompt ID is provided, the command shows the prompt editing page with the existing prompt fields pre-filled.
3) User enters the sender filter (optional) and prompt text.
3) When 'save' is clicked, the prompt is added/updated to the configuration.
