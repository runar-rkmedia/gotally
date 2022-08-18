# Tinkering with [golive](https://github.com/jfyne/live)

For a while now, I've been working a bit with off-mainstream techniques to
common web-problems. The goal has been to gain a bit more experience with other
ways of working, expand my mind, keep things interesting and find a better way
to get stuff done.

So for this project, I thought I would use golive to create a gameserver for
TallyBoard, where the client simply sends the button-presses, selections and
such to the server, and the server sends back the updated html for the
game-content.

> Hold on, isn't a game a terrible fit for server-side-rendering?

Yup, it most cases it does not make much tense, unless the game is more
text-centric or otherwise very static. A platformer would probably not work.

TallyBoard sits somewhere in between. Fast reflexes are not required, and there
are not many animations. I get to keep all the logic on the server, while the
frontend is really _dumb_.

It also lets me prototype fairly quickly.

I must confess though, that this will most likely not be the final product.

# Initial experience

Golive seems like a great kind of technology, although the project is not at
all ready for larger projects.

For smaller projects that don't require too much client-side logic, I would
like to revisit this technology, or prehaps even look into using [Phonix
LiveViews](https://github.com/phoenixframework/phoenix_live_view) which is the
project that golive is deeply inspired by, even though I have never tried
Elixir.

I really like how fast I can get interactive pages without writing JavaScript,
and all the complexities that involves.

The basic game was rendered and working nicely within half an hour, although
without animations and only the bare minimum of styling.

![Early version of the game](./images/cleary_more_hints.png)

I added some sliding-animations, and for now, I am done. It works great for
prototyping, but for production a better solution is needed.

# Problems

Payloads which rerender often can accumulate to a larger binary-size than using
a regular SPA with an api for serving data, since 
golive sends the updated html, even though it only sends the parts that are changed.

Adding client-side logic can be a bit weird without hooking into their
npm-package and creating a js-build-step. For this project, I wanted to add
some animations.





