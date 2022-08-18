# Improving performance

Today I had a go at the Hint-system, which is also used to check if the player
has no more valid movies. I had a bug where the implementation did not find all
the bugs, which is critical since the system would then decide the user was
Game Over, even when there are obviously more valid moves.

I has earlier simply [attached a board to one of the Unit
tests](https://github.com/runar-rkmedia/gotally/blob/faed4e41c998c0301ea16b5495a7edd347a875fa/tallylogic/hint_test.go#L40-L61),
which renders into a board like below:

![There clearly are more hints here, right?](./images/clearly_more_hints.png)

Fixing the bug itself wasn't all that interesting, but it revealed that I
haven't tested the speed of getting hints for boards with lots of solutions, or
how fast it is.

So I added a very stupid board to the test, with only ones everywhere on 5x5
grid. Since 1 times 1 = 1 and 1 times 1 times 1 = 1 and so on, this should
produce a lot of solutions. It turns out about 16000. 

This ran just crazy slow, expected about 260 hours.

To me, this really shows how important it is to use unit-testing and benchmarking. I caught this early, before the server would grind to a halt. I know this function will run a lot, especially since I plan to add a board-generator which calculates all the possible moves down the road.

# Reducing hint-retrieval-time by from 260 hours to two seconds.

I had a big laugh upon seeing this, and then checked why it was slow. The calculation of all the hints was 
took about 2s here, which of course is too much, but for an early concept I thought it was ok. I knew I could 
improve the performance later when writing the code.

The problem was there were some terrible copying going on simply to deduplicate some of the hints,
since the code above would produce hints that were in all fairness equal, just in reverse.

What I had done here was to loop over the whore list of hints, create a canonical representation of the hint,
then compare that to every other hint in the list. That won't scale.

So I made a quick rewrite, and made the hints include a hash of the hint for comparison and used that in a hashmap.
Now it is just the 2 seconds.

# Further benchmarking

I wrote a [quick benchmark](https://github.com/runar-rkmedia/gotally/blob/efa382f9f1a079a75c3ed3738dfa45419c0ec6a8/tallylogic/hint_test.go#L124-L146) for the same board:

```golang
func BenchmarkGetHints5x5ofOnes(b *testing.B) {
	board := TableBoard{
		cells: cellCreator(
			1, 1, 1, 1, 1,
			1, 1, 1, 1, 1,
			1, 1, 1, 1, 1,
			1, 1, 1, 1, 1,
			1, 1, 1, 1, 1,
		),
		rows:    5,
		columns: 5,
	}
	g := &hintCalculator{
		CellRetriever:      &board,
		NeighbourRetriever: board,
		Evaluator:          board,
	}

	for i := 0; i < b.N; i++ {
		g.GetHints()
	}

}
```

and yep, it still runs a bit slow:

```
goos: linux
goarch: amd64
pkg: github.com/runar-rkmedia/gotally/tallylogic
cpu: 12th Gen Intel(R) Core(TM) i9-12900H
BenchmarkGetHints5x5ofOnes-20    	       1	2023108603 ns/op	1851626032 B/op	17831818 allocs/op
BenchmarkGetHints3x3ofOnes-20    	    5768	    178542 ns/op	  136553 B/op	    3245 allocs/op
PASS
ok  	github.com/runar-rkmedia/gotally/tallylogic	3.080s
```

Since I now have a baseline, I can start chugging away at this.

Since the getHints works by calculating all the possible paths one can take from a starting-brick,
and this function is called for each of the bricks on the board, I thought it was natural to use a worker for each of these function-calls.

For this board, that meant 25 channels.

Putting this into benchstat got me these results:


```shell-script
name                  old time/op    new time/op    delta
GetHints5x5ofOnes-20     2.02s ± 0%     0.79s ± 0%   ~     (p=1.000 n=1+1)
GetHints3x3ofOnes-20     179µs ± 0%     139µs ± 0%   ~     (p=1.000 n=1+1)

name                  old alloc/op   new alloc/op   delta
GetHints5x5ofOnes-20    1.85GB ± 0%    1.86GB ± 0%   ~     (p=1.000 n=1+1)
GetHints3x3ofOnes-20     137kB ± 0%     138kB ± 0%   ~     (p=1.000 n=1+1)

name                  old allocs/op  new allocs/op  delta
GetHints5x5ofOnes-20     17.8M ± 0%     17.9M ± 0%   ~     (p=1.000 n=1+1)
GetHints3x3ofOnes-20     3.25k ± 0%     3.27k ± 0%   ~     (p=1.000 n=1+1)
```

Nice, it runs now for **0.79s** compared to **2.02s**. I don't get a delta-value directly here, since the sample-size is too small.

Still, there are room for improvements, especially with the allocations.

