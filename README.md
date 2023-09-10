# JIM: A Simple Chat System
*(This code is for learning purposes only and is strictly prohibited for illegal activities)*

*WARNING:*
    This readme document has been translated from the Chinese version using ChatGPT. If you can read Chinese, please visit this link: [[中文版]](./README_zh.md).

![](https://img.shields.io/github/actions/workflow/status/jerbe/jim/cross-build.yml)
![](https://img.shields.io/github/issues/jerbe/jim?color=green)
![](https://img.shields.io/github/stars/jerbe/jim?color=yellow)
![](https://img.shields.io/github/forks/jerbe/jim?color=orange)
![](https://img.shields.io/github/license/jerbe/jim?color=ff69b4)
![](https://img.shields.io/github/languages/count/jerbe/jim)
![](https://img.shields.io/github/languages/code-size/jerbe/jim?color=blueviolet)


## Introduction
JIM (Jerbe's Instant Messaging) is a lightweight chat system with the following features

* Written purely in Golang, supporting private chat, group chat, and world channel chat.
* Implements some common functionalities, including friend invitation, friend search, friend addition, friend deletion, group creation, group joining, group invitation, and group exit.
* Easily extensible business logic with relatively clean code and detailed comments.
* A relatively complete logging system that records various business fields in detail. Logs are in JSON format, suitable for ELK collection and analysis.
* Supports cross-platform compilation and deployment using cross-compilation.
* Uses the Swag documentation specification, and you can generate the corresponding API documentation with swag init, making it easy for front-end developers to debug.
* ~~Beautiful front-end interface~~
* Utilizes GitHub Actions for multi-platform compilation, and you can download corresponding platform executables from releases.

## Related Components
    Language:  golang  
    Dependencies:  redis、mongodb、mysql/mariadb
    Logging package： github.com/rs/zerolog
    SQL operations: github.com/jmoiron/sqlx

## Design
    1) MySQL is used as the storage database for user information, friend relationships, and group information. MongoDB is used as the message storage database. Redis is used as the subscription service and caching storage service.
    2) Users obtain tokens upon login and use tokens to establish WebSocket connections. Since sending messages requires authentication and some policy filtering, WebSocket connections are generally only used to receive server-side messages, and sending chat messages does not depend on WebSocket.
    3) In the case of multiple service instances, since the server that a user establishes a WebSocket connection with is random, we use the manager of the `websocket.manager` package in each service instance for unified management. Subscribed data is ultimately distributed to different user connections by the manager.
    4) Because we need to manage chat messages, when sending a message, the message is first stored in the database and then subscribed and pushed after being successfully stored.
    5) Supports getting the latest X messages from each room and traversing historical messages.

## Design Diagrams


##### Basic Architecture Diagram for Message Sending

![](./assets/聊天架构设计.jpg)

##### Message Sending Sequence Diagram

![](./assets/时序图.jpeg)

## Initialization
    1. Deploy the relevant services: MySQL/MariaDB, MongoDB, Redis.
    2. Configure the `config.yml` file from the `config` folder with the corresponding service addresses and modify them.
    3. Import the SQL statement files from the `sql` folder into MySQL/MariaDB. `all.sql` contains all the database creation statements, while the others contain individual table creation statements.


## Features
Here are the planned features:

### Accounts
+ [x] Websocket Channel
  - [x] Cross-server push
  - [x] Receive chat messages
  - [x] Receive notification messages
+ [x] User Registration
  - [x] Verification code validation
  - [ ] Strong password validation
  - [ ] Email binding
  - [ ] Limited registration time switch
+ [x] User Login
  - [x] Verification code validation
  - [x] Cannot log in if disabled
  - [x] Cannot log in if deleted
- [ ] User Logout
- [x] Account Information
  - [ ] Password modification
  - [ ] Avatar modification
  - [ ] Nickname modification
  - [ ] Online status modification

### Friends
- [x] Find Friends
  - [x] Search by ID
  - [x] Search by nickname
- [x] Friend Requests
  - [x] Add remarks when requesting friends
  - [x] Greet when becoming friends
- [x] Friend Editing
  - [x] Modify remarks
  - [x] Remove friends
  - [ ] Delete all chat records after removal
  - [x] Block friends
  - [ ] Cannot be added as friends again after blocking
- [ ] Friends List
  - [ ] Invisible friends not displayed

### Groups
+ [x] Create Groups
+ [x] Join Groups
+ [x] Group Member Management
  - [x] Limit the number of members
  - [ ] Blacklist group members cannot join again
  - [x] Remove group members
  - [ ] Set blacklists when removing group members
  - [x] Appoint group administrators
  - [ ] Modify group member nicknames
+ [x] Modify Group Information
  - [x] Transfer group ownership
  - [x] Modify group nickname
  - [ ] Modify group avatar
- [x] Exit Groups
- [ ] Dissolve Groups

### Chat
- [ ] Chat List
  - [ ] Pin chats
  - [x] Last message in chat room
- [x]  Private Chat
  - [x] Send plain text
  - [ ] Send images
  - [ ] Send emoji
  - [ ] Send videos
  - [ ] Send voice
  - [ ] Send location
  - [ ] Voice chat
  - [ ] Video chat
  - [ ] Mark messages as read
  - [x] Cannot chat if blocked by a friend
  - [x] Cannot chat if deleted by a friend
- [x] Group Chat
  - [x] Send plain text
  - [ ] Send plain text
  - [ ] Send images
  - [ ] Send emoji
  - [ ] Send videos
  - [ ] Send voice
  - [ ] Send location
  - [ ] Voice chat
  - [ ] Mark messages as read
  - [x] Mute all
  - [x] Mute specific group members
- [x] World Channel Chat
  - [ ] World channel switch
  - [ ] World channel mute

# Other
[[API Documentation]](http://github.com/jerbe/jim-docs)，[[Jcache - (Encapsulated Distributed Cache Integration Solution)]](http://github.com/jerbe/jcache)
  