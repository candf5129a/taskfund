
---

# Dashboard Layout

```text
┌────────────────────────────────────────────────────────────┐
│ Logo               Search              🔔      👤          │
├───────────────┬────────────────────────────────────────────┤
│ Dashboard     │ Wallet Card                               │
│ Tasks         ├────────────────────────────────────────────┤
│ Wallet        │ Statistics Cards                          │
│ Analytics     ├────────────────────────────────────────────┤
│ Notifications │ Active Tasks                              │
│ Support       ├────────────────────────────────────────────┤
│ Settings      │ Available Tasks                           │
│               ├────────────────────────────────────────────┤
│               │ Recent Activity                           │
└───────────────┴────────────────────────────────────────────┘
```

---

# Sidebar

```text
🏠 Dashboard

📋 Tasks

💰 Wallet

📊 Analytics

🔔 Notifications

❓ Support

⚙ Settings
```

Simple.

No clutter.

---

# Top Navigation

```text
Logo

Search

Notifications

Messages (Future)

Profile
```

---

# Wallet Card

Large card at the top.

```text
Available Balance

₦12,450

Today's Earnings

₦850

Withdraw
```

---

# Statistics Cards

Four cards.

```text
Available Tasks

145
```

---

```text
Active Tasks

4
```

---

```text
Approval Rate

98%
```

---

```text
Completed Tasks

1,245
```

---

# Active Tasks

```text
TikTok Like

₦500

Time Left

15 mins

Continue
```

---

# Available Tasks

Card example.

```text
Install App

₦800

★★★★☆

340/500 Slots

02:15 Remaining

View
```

---

# Notifications

Only show important updates.

```text
Withdrawal Approved

Task Approved

New Campaign Available
```

---

# Quick Actions

A small section.

```text
Browse Tasks

Withdraw

Edit Profile

Support
```

---

# Footer

```text
Privacy

Terms

Help

Version 1.0
```

---

# Design Rules

Document these inside the dashboard file.

* Maximum 3 colors on one screen.
* One primary button per section.
* Cards have equal height.
* Icons always appear before labels.
* Keep important information above the fold.

---

# Dashboard Components

List every component used.

```text
Sidebar

Navbar

Wallet Card

Statistic Card

Task Card

Notification Card

Buttons

Search Bar

Avatar

Progress Bar
```

These become reusable throughout the app.

---

# Dashboard States

Every page should have these states planned.

### Empty

No tasks available.

---

### Loading

Skeleton loaders.

---

### Error

Unable to load tasks.

---

### Success

Everything displayed normally.

---

# Responsive Layout

Desktop

```text
Sidebar

+

Main Content
```

Tablet

```text
Collapsible Sidebar
```

Mobile

```text
Bottom Navigation
```

---

# 🚀 After Dashboard

Once this document is complete, our design order will be:

```text
✅ Worker Dashboard

↓

Browse Tasks

↓

Task Details

↓

Wallet

↓

Profile

↓

Notifications

↓

Advertiser Dashboard

↓

Campaign Pages

↓

Admin Dashboard

↓

Landing Page
```

---

# ⭐ One Improvement to Our Process

From this point onward, **every page should follow the same template**. Create a reusable page design checklist.

For every screen, document:

```text
Page Name

Purpose

Target User

Components Used

Actions Available

API Endpoints Used

Database Tables Used

Possible States
    • Loading
    • Empty
    • Success
    • Error

Responsive Behavior

Accessibility Notes
```

For example, for the **Worker Dashboard**:

* **Purpose:** Give workers a complete overview of their account and tasks.
* **Target User:** Worker.
* **Components:** Sidebar, Navbar, Wallet Card, Statistic Cards, Task Cards, Notifications.
* **API Endpoints:** `/api/v1/wallet`, `/api/v1/tasks`, `/api/v1/notifications`.
* **Database Tables:** `users`, `wallets`, `tasks`, `task_submissions`, `notifications`.

Using the same template for every page will make both the design phase and the development phase much more organized and predictable.
