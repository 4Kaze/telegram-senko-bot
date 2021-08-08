# Telegram-senko-bot


[![MIT License][license-shield]][license-url]
[![Chat on Telegram][telegram-bot-shield]][telegram-bot-url]
![build-shield]

<p align="center">
  <img src="https://user-images.githubusercontent.com/47448124/128641986-732215c0-2d7e-4795-9817-20410a0dd212.gif" alt="UwU! An example!"/>
</p>

## About The Bot

This bot greets new members on telegram groups with a cute, personalised gif with Senko-san.

In case you're wondering, Senko-san is an adorable 800-year old fox goddess from the anime [The Helpful Fox Senko-san](https://myanimelist.net/anime/38759/Sewayaki_Kitsune_no_Senko-san) (jp. *世話やきキツネの仙狐さん*).

<!-- GETTING STARTED -->
## Getting Started

Follow these steps to set up your own version of this bot.

### Prerequisites

* Go 1.13 (see [instructions](https://golang.org/doc/install))
* FFmpeg
  ```sh
  sudo apt install ffmpeg -y
  ```

### Set up using Google Cloud Functions

1. Download and install the Google Cloud SDK, for example [with Snap](https://cloud.google.com/sdk/docs/downloads-snap):
   ```sh
   snap install google-cloud-sdk --classic
   gcloud init
   ```
2. Clone the repo:
   ```sh
   git clone https://github.com/4Kaze/telegram-senko-bot.git
   ```
3. Download [**NotoSansCJKjp-Black.otf**](https://www.google.com/get/noto/help/cjk/) font into the project folder:
    ```sh
   wget https://noto-website-2.storage.googleapis.com/pkgs/NotoSansCJKjp-hinted.zip -O fonts.zip
   unzip fonts.zip NotoSansCJKjp-Black.otf
   rm fonts.zip
   ```
4. Deploy the bot to Google Cloud Functions replacing **[YOUR TOKEN]** with your bot token from [BotFather](https://t.me/botfather):
    ```sh
   gcloud functions deploy SenkoSan --entry-point=HandleRequest --runtime go113 --trigger-http --allow-unauthenticated --set-env-vars TOKEN=[YOUR TOKEN]
   ```
5. If everything went well, there should be a trigger url in the output in your terminal:
    ```
    HttpsTrigger:
        url: https://us-central1-bot-tele-002137.cloudfunctions.net/SenkoSan
    ```
   Set this url as a webhook for your bot:
     ```sh
   curl https://api.telegram.org/bot[YOUR TOKEN]/setWebhook?url=[YOUR URL]
      ```

### Usage

The bot simply sends a gif with new member's name when they join a group.

In private messages, it supports the following commands:
* `/start` - a standard command that displays the description
* `/genewate [name]` - a command that generates a new gif with given name 
* `/wepo` - a command that sends the link to this repo

All names are stripped from emojis and truncated to maximum 20 characters.

## Contributing

I'm open for contributions. Feel free to fork the project to use it for your own means or to create a PR with new features / fixes.

## License

Distributed under the MIT License. See `LICENSE` for more information.


<!-- CONTACT -->
## Contact

[![Chat on Telegram][telegram-shield]][telegram-bot-url]


## Acknowledgements

* [Readme template](https://github.com/othneildrew/Best-README-Template)
* [OwO-ifier](https://lingojam.com/OwO-ifier%28UwU%29)


[license-shield]: https://img.shields.io/github/license/4kaze/telegram-senko-bot
[license-url]: https://github.com/4Kaze/telegram-senko-bot/blob/main/LICENSE
[telegram-bot-url]: https://t.me/joinchat/DYKH-0G_8hsDDoN_iE8ZlA
[build-shield]: https://img.shields.io/github/workflow/status/4Kaze/telegram-senko-bot/Go
[telegram-bot-shield]: https://img.shields.io/badge/Demo-Senko-green?logo=telegram
[telegram-shield]: https://img.shields.io/badge/-Contact%20me%20on%20Telegram-gray?logo=telegram
[telegram-profile-url]: https://t.me/yonkaze
