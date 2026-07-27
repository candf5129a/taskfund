# Our Database Structure

### This is what I recommend for Version 1.
```
Users

Roles

Permissions

Profiles

Tasks

TaskCategories

TaskSubmissions

Campaigns

Wallets

Transactions

Withdrawals

Notifications

Reviews

Ratings

SupportTickets

Settings

AuditLogs
```


## Example
### Users:
```
id

first_name

last_name

username

email

phone

password_hash

role_id

status

created_at

updated_at
```
##
### Tasks:
```
id

campaign_id

title

description

reward

slots

slots_remaining

approval_rate

status

expires_at

created_at
```
##
### Wallet:
```
id

user_id

balance

pending_balance

total_earned

total_withdrawn
```
##
### Transactions:
```
id

wallet_id

amount

type

status

created_at
```
##
### Reviews:
```md
id

reviewer_id

reviewed_user_id

rating

comment

created_at
```
##
### Relationships

Document every relationship.

### Example:
```md
User

↓

One Wallet

↓

Many Transactions

↓

Many Withdrawals
```
```md
Advertiser

↓

Many Campaigns

↓

Many Tasks

↓

Many Submissions

↓

Many Workers
```