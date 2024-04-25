![Decanter Logo](https://github.com/p5quared/decanter/assets/98245483/64732713-5950-40ad-bd8f-6fb9e5b06d84)
# Decanter
Decanter is a CLI app for interacting with Autolab at UB (University at Buffalo ...).
_Pour_ data from your computer to Autolab, a little more seamlessly.

## Installation

If you have [homebrew](https://brew.sh/) you can install via tap like so:
```shell
brew install p5quared/decanter/decanter
```
Otherwise, you can build from source if you have [go](https://go.dev/) installed:
```shell
git clone https://github.com/p5quared/decanter.git
cd decanter
go install
```

## Usage

Before you can do much of anything, you need to setup Decanter via `decanter setup`.

To view a command overview, use `decanter -h`.

Some examples are:

* `decanter list me`
* `decanter list assessments`
* `decanter submit -c cse305-s24 -a pangram -f main.ml`

## Tips

Remembering the full submit command can get quite tedious
for your projects. Consider writing a per-project script
to automate this process. For instance, if your project
came with a `Makefile` or you want to use 
[just](https://github.com/casey/just), you could 
add a recipe like so:

```makefile
submit: submission
	decanter submit -c COURSE_ID -a ASSESSMENT -f FILE -w 
```

## Notes

If you find issues or have suggestions, 
I'd love to hear about them in the Github issues section.

Contributions would be appreciated. If you'd like to contribute,
please email me or write an issue first. There's also a long
list of TODO's that I have written down and scattered throughout
the code.

At some point I'll be cleaning up the Autolab API and publishing
it as a standalone library.

If you've read this far I'd really appreciate if you'd give the repo
a ⭐ or a watch. We need to reach a certain threshold of 'notoriety' 
before Homebrew will accept our formulae.

## FAQ

* _When I view my submission scores, I can't see the total (score / x)!_
    * The Autolab API only returns problem scores, and requires instructor scopes to access maximum scores. *shrug*
* _Why can't I do XYZ with Decanter?_
    * Hey come on, I'm only one person here.
* _Why is it red and not blue?_
    * Red is the color of passion, here are some great red things:
        * The Northern Cardinal
        * The <3 in I <3 NY
        * The leaves of a Japanese maple tree.
        * My copy of _Crime and Punishment_
        * _etc..._
    * If that doesn't satisfy you, I also want make sure nobody is confused about the fact that this app is not directly affiliated with UB in any way.
* _What the hell is that logo?_
   * Sorry honey, it's modern art and I'm afraid _you just don't get it._
