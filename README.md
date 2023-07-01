# Gmail to Telegram Bot
This bot is designed to forward messages from Gmail to Telegram. It periodically checks for new emails in a specified Gmail account and sends the content of those emails, including the timestamp, subject, and body, to the specified Telegram user.

Additionally, this bot can be used to send emails by simply typing a message in a specific format. The message should be formatted as follows:
`recipient_email ; subject ; body`
The recipient email, subject, and body should be separated by a semicolon and a space.

## Installation
Clone this repository to your local machine:

`git clone https://github.com/your_username/gmail-telegram-bot.git`

`cd mail-sender`
`go get .`

Before running the bot, you need to configure your Gmail and Telegram API credentials.

## Gmail Configuration
-- 1. Enable the Gmail API and create a new project in the Google Cloud Console. <br>
-- 2. Set up OAuth 2.0 credentials and download the JSON file containing your client ID and client secret. <br>
-- 3. Rename the downloaded JSON file to credentials.json and place it in the config directory. <br>


## Telegram Configuration

-- 1. Create a new bot on Telegram by talking to the BotFather.
-- 2. Obtain the bot token and save it for later use.

## Bot Configuration
gmail.username: Your Gmail email address.
gmail.accessTokenPath: The file path where the Gmail access token should be stored.
telegram.botToken: The token of your Telegram bot.
telegram.chatId: The chat ID of the Telegram user you want to send messages to.
Usage
To start the bot, run the following command:

bash
Copy code
npm start
The bot will check for new emails every 5 minutes and forward any new messages to the specified Telegram user.

To send an email using the bot, simply send a message in the following format to your Telegram bot:

css
Copy code
recipient_email ; subject ; body
Make sure to replace recipient_email, subject, and body with the actual values.

Contributing
Contributions are welcome! If you find any issues or have suggestions for improvement, please open an issue or submit a pull request.

License
This project is licensed under the MIT License.
