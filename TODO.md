# Todos

## Important (Before Release)
These are mostly bigger tasks that will take several
hours to complete.

### Configure Homebrew Formulae
This will take a while.

### Final Visual
* Viewing submissions

## Backlog
Mostly trivial; could be done by anyone.

### Asessments.txt
Would be nice to not have to specify submission
info every time


### Publish Autolab Client
Consider publishing Autolab Go Client.
Possibly useful as Autolab does not implement
oAuth2 according to official standards.

### Visual
* Formalize design
* Form themes
* Global lipgloss elements

### Refactor
* Modules for:
    * interaction with proxy server
    * visual elements
* Create generic wrapper for server interactions
    * i.e. doWithSpinner(loading, done string, func () {})

### Fuctional
* Testing
    * Mock Autolab OAuth Server
    * Mock Autolab API
    * Mock proxy server
* Caching
    * At the moment we are not caching anything.
    Total requests could probably be greatly reduced
    by caching data like courses which shouldn't change
    often at all.

### Interactive mode?
* Is it worthwhile to create an interactive mode with
charmbracelete/bubbletea?

### Open in Browser
Might be nice to be able to open submissions in browser.
