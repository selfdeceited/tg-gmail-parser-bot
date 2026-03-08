# Start
1) User registers in the bot and sends /start. It starts the installation guide telling how to set up GCP OAuth token for email access. `/register` command is available.

# Register
this command is used to link with new Gmail account for monitoring.
1) User sends the `/register` command. Command shows the installation guide once again and asks the user to paste the credentials json result from GCP OAuth.
2) The bot creates the refresh + access token to perform the smoke test (can read last email and verify that the content is not empty)
3) if the smoke test fails, the bot tells the user that credentials are invalid.
4) if the smoke test passes, the bot saves the user data with telegram user id and the encrypted refresh token.
5) The bot validates the token is working and saves the GCP OAuth token for email access.
6) The 'configure' button is available below using [Special Keyboard](https://core.telegram.org/bots/features#keyboards) as well as `/configure` command.
7) If user clicks on 'configure' button with credentials present, verify them again.
- if they are still valid, show that the registration is active. If you want to re-register, visit /clearregistration command.
- if they are invalid, clear credentials and start over the registration process.

# Clear Registration
this command is used to clear the registration and start over. It's not available if `register` command result is not successful.
1) User sends `/clearregistration` command. Command clears the registration data and shows the installation guide again.

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
