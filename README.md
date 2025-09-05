# Consensus

Scrum poker site with a compact UI.

Status: MVP - Basic functionality implemented. Needs further refinement for UI/UX and code improvements.

## Features

- AuthZ/AuthN deferred to [OAuth2-Proxy](https://github.com/oauth2-proxy/oauth2-proxy)
- Add tickets and link to the ticket
- Point based on fibonacci sequence
- Show who has/hasn't voted
- Show average/most frequent votes
- Compact UI

## TODO

Things to complete before changing the project status.

### Alpha

Basic functionality that we are missing.

- [ ] Persistence. Likely SQLite.
- [ ] Display errors in UI
- [ ] Ability to remove Tickets
- [x] "Live" view so tickets update without refreshing

### Beta

Some extra features that we should have.

- [ ] Customisation of voting values
- [ ] "Question" voting value
- [ ] Refine error handling
- [ ] Basic functionality tests
- [ ] Edit tickets

### Stable

- [ ] UI/UX refinement
  - [ ] Smoother "live" view
  - [ ] Ensure consistent order of tickets and other items
- [ ] URL validation
- [ ] Code cleanup (e.g. TODO items generally removed, idiomatic go, etc)

### Unsure

- [ ] Multi-tenancy
- [ ] Integration with external ticket systems (e.g. Jira)
