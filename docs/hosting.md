# Creating and adding the bot to your server
## Creating the bot
- Head over to the [Discord Developer Portal](https://discord.com/developers/applications)
- Create a new application by clicking the **New Application** button on the top right corner

## Obtaining the token
- Select your newly created application, and go to the **Bot** tab
- Click on **Reset Token**, and copy and paste the shown token on your `config.yml`
  Note:
- On this page you can also change the username and update the icon for your bot
- If you plan to let others add your bot, enable the **Public bot** toggle

## Adding the bot
- We need to generate the URL: click on **OAuth2** -> **URL Generator**
- For the scopes tick: **bot** and **applications.commands**
- For the permissions: Send **Messages**, **Connect** and **Speak**


![Screenshot 2023-08-13 at 16-08-04 Discord Developer Portal — API Docs for Bots and Developers](https://github.com/TheTipo01/YADMB/assets/10187614/94859744-285d-4130-b526-a4ea3f49994a)


- On the bottom of the page, you will find the generated URL: with that, you can finally add your bot!
