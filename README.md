# stakeclaim

A simple utility to claim vote rewards and refresh vote strength daily on the WAX blockchain.

> [!IMPORTANT]
>
> _This utility can **only** be used to delegate the votes to a proxy account, it cannot be used to vote for individual block producers_

## Work

This utility checks if the last claim time is less than 24 hours, if yes, it waits until the next claim time. otherwise it checks if the account has voted before, if yes then it claims the rewards and refreshes the vote strength, otherwise it just votes.

This is an image of a flowchart shown below

![Flowchart](https://mermaid.ink/img/pako:eNpdkcFOwkAQhl9lMgdTkkJsgUp7MBEIcBAvNSZqPazbqW1od8l2qyLw7g4tB2UPm5l_v_lnd2ePUqeEEWal_pK5MBYe54kCXneOE1sWej3o929h-jrLSW6gFLUFWYqiAltU9NbB05Y53FNdg82FAn8EK92Y-gCxE5dE294_cK0NXYILp-tQZCCk1I2y8KktpfBOGeNngwUbnByeiUuWzqy9yRU8MXlBPOgDrJw_B6u2d9wly79J3D2xSxKFLlZkKlGk_DP7k5qgzamiBCMOU2E2CSbqyJxorI53SmJkTUMuNttUWJoX4sOICqNMlDWrW6Ew2uM3Rp4fDnw_mEy8MBwNh-NR6OKO5WDgBdfj8GYcBH4Q8n508UdrtvDa8pc2PttRWlht1t3k2gEefwFAiIZW?type=png)

## Usage

Before using this utility, you need to have a config file.

By default, it will be look for a file named `config.txt` in the current working directory.

The syntax of the config file is as shown below, 1 account per line, you can add as many accounts as you want.

```
account:permission:private_key:proxy
# blank lines and lines starting with # are considered as comments, they are ignored
```

Sample config.txt file

```
# this is an example config file
anchor.gm:claim:5K*************************************************:top21.oig
mycoolwallet:vote:5K*************************************************:waxcommunity
hello.wam:active:5K*************************************************:bloksioproxy
```

> [!NOTE]
>
> If you want to use this with a [Cloud Wallet](https://mycloudwallet.com) account _(.wam account)_, you will need to soft/hard claim your account first.

> [!WARNING]
>
> It is **highly** recommended that you create a custom permission that can only do the `eosio::voteproducer` and `eosio::claimgbmvote` actions for increased security.
>
> You can read more about custom permissions in this [article](https://academy.anyo.io/eosio-private-key-system/) by Anyobservation Academy, and how to make custom permissions in this [course](https://waxsweden.org/course/adding-custom-permission-using-bloks/) from WAX sw/eden

### CLI

Download the latest [release](https://github.com/benjiewheeler/stakeclaim/releases/latest), and run it in a terminal using the following command:

```sh
./stakeclaim
# Or you can pass a path to a config file
./stakeclaim --config-file /path/to/config.txt
```

### Docker

When running the utility in Docker, you must create a config file named `config.txt` and mount it to the container under the `/config` folder.

```sh
# start the container
docker run -d --name=autoclaim -v /path/to/config.txt:/config/config.txt ghcr.io/benjiewheeler/stakeclaim:latest
```

## Want more features ?

_Hire me_ ;)

[![Discord Badge](https://img.shields.io/static/v1?message=Discord&label=benjie_wh&style=flat&logo=discord&color=7289da&logoColor=7289da)](https://discordapp.com/users/789556474002014219)
[![Telegram Badge](https://img.shields.io/static/v1?message=Telegram&label=benjie_wh&style=flat&logo=telegram&color=229ED9)](https://t.me/benjie_wh)
[![Protonmail Badge](https://img.shields.io/static/v1?message=Email&label=ProtonMail&style=flat&logo=protonmail&color=6d4aff&logoColor=white)](mailto:benjiewheeler@protonmail.com)
[![Github Badge](https://img.shields.io/static/v1?message=Github&label=benjiewheeler&style=flat&logo=github&color=171515)](https://github.com/benjiewheeler)

## License

Licensed under the [MIT License](LICENSE)
