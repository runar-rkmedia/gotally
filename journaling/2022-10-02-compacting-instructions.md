# Compacting instructions

This post is about how I've been trying to make the instruction, or the history of a Tally Board game take up less space.

Games of Tally Board can theoretically consist of an almost infinite amount of moves. In practice, most games are short, with less than 100 moves, but some can be more than 10000.

To be able to view a replay of a game, I like to record them, and therefore I must store every swipe, combining of tiles and other sorts of moves.

A swipe is easy, just a simple enum, and it should be fairly compact already. The problem is when I am trying to record the combining of tiles. These technically consist of an array of ints, where the length of the array is fixed to the size of the board, and the insts as well are limited to this size. So for a board of size `5x5`, a `[25]uint8` seems at first glance like a good idea.

However, this does not scale very well.

## Pre-generated maps

A generated map of every combination can be generated and sorted in a stable way and then inserted into a `map[][]uint8` which can then be used as a reference in the code. This is fairly straight-forward, so I took a swing at it, and then used the `gob` (go-binary-format) to marshall it disk, and then read that to memory on startup. 

The pregenerating step takes about 10 seconds or so, and results in this map, and an additional one for reverse-lookup (in the form of a tree). Both maps contains a bit short 4 million items. The problem is, this raises the memory-consumption by too much (from 5 MB 227 MB for the initial map). This can probably be probably be reduced, but when writing this article I remembered a very different way to reduce the size of ints. 

## Reducing the size of ints slices.

I TimeSeries-databases, they do this trick where they compress the numbers really compactly. Although they mostly work with floats here, some of the same tricks should be possible here.

Basically, this boils down to something along the lines of:

- Combining smaller ints into one or more `int64`'s.
- Diffing the previous int in the slice with the next. This makes it possible to combine many more of the ints into a single `int64`.- Diff of diffs. This takes the previous step a bit further

### Constraints to the rescue

In Tally Board, there are some rules for what tiles can be combined.

1. It is only possible to select tiles that are not empty.
2. **A subsequent selection can only select neighbours of the previous selection.**
3. Tiles cannot repeat in the path.
4. The last item selected must have a value greater than the rest, and must evaluate for that path.

Especially the second point here is interesting, since that means that after the first selection, at most four choices for the second selection, and then there can never be more than three choices for the rest. Often there are only two or even just a single choice. That means that we can pack a whole lot of items into just this information. For a 5x5 board, there is 3.060.392 possible choices total, in the worst case scenario.

## Restating the problem 

Could we calculate both lookup and reverse-lookup at runtime for the values we need? That removes the need for the pregenerated maps, and the memory footprint should be almost gone.

Going from a path to this reduced value can be done with a simple hash-function. The only requirement is that it has no crashes between values, and is within the range of a single uint64 or better.

With all of 64 bits available, I could probably just store the first index a pure integer, and the following integers just as simple choices within the rest of the available bits as flags.

The first 6 bytes is all I need for the index, so I have 54 bytes left for the flags.

Since we only need the path when the game is loaded, and we are at the correct state for the path to apply, we can reduce this further with the additional Constraints listed above, but I don't want to overcomplicate things unless I need to.

This way is fairly easy to implement, and can be reversed. It will require a bit of bit-shifting, Though. I haven't done that in quite some time, so I will need to read up a bit on that.

## Using what we have

For any given point in the game, we already have a method that tells us every possible path that can be combined, and this will return a hash of the path. We can use this hash-implementation for the reverse-lookup, and probably for the lookup too.

However, this hash is a string. Perhaps it should have used the bitpacking mentioned above instead.

## KISS

Why even bother with this? Just store the path in whatever format makes sense, even a concatenated string, with a separator would work. We probably do not need the history very often, and can just send it off to the database for storage. We could even use the pre-generated data from above, and put that into the database.

In any case, how many players do I really think I will have, playing how many games? I mean, I currently use PlanetScales free offering here, and that gives me plenty of storage for this, with a whopping 10 million writes per month. If I ever have problems with scaling here, I am sure going for the paid alternative is a no-brainer. Perhaps even I Tally Board is making me enough money to pay the $29 per months?





