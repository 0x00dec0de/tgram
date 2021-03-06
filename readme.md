# Typegram

[typegram](http://tgr.am) - zen blog platform (like medium).

**Basic Capabilities**

 - publications, comments
 - favorites, subscriptions
 - mentions, tags
 - ratings, votes and so on

**Playground**

You can try the service on a special [test site](http://tst.tgr.am/). Please! Use this [playground](http://tst.tgr.am/) for play with engine!

**Localization**

The service is available, for example, for [Russian-speaking](http://ru.tgr.am/), or [English-speaking](http://en.tgr.am/) users. During development, platforms for other languages are opened. On each subdomain, users and publications are separate. Please, help me to translate the welcome post for your language
[Add my country](https://github.com/recoilme/tgram/issues/43)

**Optimization**

The first thing that catches your eye is the high speed of page loads and aggressive optimization.
![](https://tst.tgr.am/i/tst/recoilme/17_.png)


You will not find third-party scripts that monitor user behavior or huge styles / images. The site works with javascript turned off, it remains fast and convenient on any platform.

**Subscriptions**

On the main page, the author you are subscribed to is displayed, and the number of new publications. The link leads to the first unread message, in chronological order, as in telegram. Typegram does not impose on you whom and when to read.
![](https://tst.tgr.am/i/tst/recoilme/23_.png)

**Mentions**

When someone mentions you in comments you will see it on the main page
![](https://tst.tgr.am/i/tst/recoilme/22_.png)

**Editor**

The editor supports typing in markdown markup, with rich features and visual formatting. With the ability to make a post fullscreen, preview, autosave and other convenient "tidbits"
![](https://en.tgr.am/i/en/recoilme/2_.png)

**Rating system**

You may see three sections with strange names on the main page:
**top mid btm**

![](http://www.wallpaperdx.com/photo/pudge-butcher-dota-abstract-art-chain-full-hd-732-416.jpg)

Yes, I love DotA (my dog's name is Pudge, for example). And I'm sure that ratings are more about game mechanics/motivation than something seriously adequate. On typegram, content is divided into three parts, top, middle and bottom. All new articles go to farm the rating on the midline. Good articles go to the top. Bad articles fall to the bottom. Technically, the ranking system is copied from the ycombinator.

**Rating of the article.**

**+ 5:1 -**

Each user has 10 votes per day. You may spend them on both pluses and minuses for one article, or distribute them as you want.

The author sees both the negative and the positive reactions, separately.

**Rating of the comments**

**+ 5**

Comments are positive only. I do not know why. Do not ask. I just want to give more opportunities for collecting feedback with different mechanics. And for comments, it is possible to give only one vote per comment. You have 10 votes for comments per day. One comment is one voice.


**Tags**

Each article may have a global tag. But only one. Be smart, then choose a tag for your article.


**Openness**

The project is free and open source. I always welcome comments and suggestions on [github](https://github.com/recoilme/tgram)

**Weekly news digest**

[subscribe on weekly digest here](https://www.producthunt.com/upcoming/typegram)

**App**

This app uses [slowpoke](https://github.com/recoilme/slowpoke) as the database. Package slowpoke implements a low-level key/value store in pure Go. This database engine was developed specially for typegram

![slowpoke](https://en.tgr.am/i/en/recoilme/3_.png)


# How it works
```
.
├── main.go
├── routers
│   ├── routers.go      //routers
└── models.go
    ├── article.go      //data models for article
    └── user.go         //data models for user

```

# Getting started


You only need golang to run typegram

## Install Golang
https://golang.org/doc/install
## Environment Config
make sure your ~/.*shrc has the following variables:
```
➜  echo $GOPATH
/Users/zitwang/test/
➜  echo $GOROOT
/usr/local/go/
➜  echo $PATH
...:/usr/local/go/bin:/Users/zitwang/test//bin:/usr/local/go//bin
```

Replace _zitwang_ with your own username.

## Fresh 

Fresh can help you rebuild and restart Typegram automatically
```
go get -u github.com/pilu/fresh
```


## Env

You may create a tgram.env file with startup params, sample:
```
TGRAMPWD=SOM2324&E*&Ff!!EDjweljf
TGRAMPORT=:8081
TGRAMTITLE=typegram
TGRAMNAME=Typegram
TGRAMDESC=zen platform for writers
TGRAMADMIN=recoilme
TGRAMABOUT=/@recoilme/1
TGRAMDOMAIN=tgr.am
```


## Start
```
➜  go get ./...
➜  go build
➜  ./tgram
```

## Thanks

[awsm.css](https://github.com/igoradamenko/awsm.css)


[realworld.io](https://realworld.io)


[inscryb-markdown-editor](https://github.com/Inscryb/inscryb-markdown-editor)


## Design

[egorabaturov](https://egorabaturov.com)

## Dev branch

- master

## Contributors

[Contributors](https://github.com/recoilme/tgram/graphs/contributors)


You are welcome to!

## Plans


I try to build the new big thing for blogging) Much more than just a text version of medium. But i started from scratch.
The mobile version will be an incredible publishing platform for both writers and readers alike. Subscribe to be the first!

https://www.producthunt.com/upcoming/typegram