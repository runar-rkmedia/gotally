package tallylogic

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

type bruteBreadthSolver struct {
	SolveOptions
}

// NOTE: The breadth-first solver is TERRIBLE at solving infinite games.
// The depth-first-solver should be used for those cases, as it is more likely
// to produce solutions with a high score.
func NewBruteBreadthSolver(options SolveOptions) bruteBreadthSolver {
	if options.MaxDepth == 0 {
		options.MaxDepth = 1_000
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 10_000
	}
	if options.MaxTime == 0 {
		options.MaxTime = 10 * time.Second
	}
	if options.InfiniteGameMaxScoreIncrease == 0 {
		options.InfiniteGameMaxScoreIncrease = 1000
	}
	return bruteBreadthSolver{
		SolveOptions: options,
	}
}

type gameJob struct {
	Game
	hash  string
	depth int
	kind  string
}

type jobs struct {
	queue map[int][]Game
	sync.RWMutex
}

func (j *jobs) get(depth int) *Game {
	j.Lock()
	defer j.Unlock()
	l := j.queue[depth]
	if l == nil {
		return nil
	}
	if len(l) == 0 {
		return nil
	}
	g := l[len(l)-1]
	j.queue[depth] = l[:len(l)-1]
	return &g
}
func (j *jobs) has(depth int) bool {
	j.RLock()
	defer j.RUnlock()
	l := j.queue[depth]
	if l == nil {
		return false
	}
	if len(l) == 0 {
		return false
	}
	return true
}

func (b *bruteBreadthSolver) SolveGame(g Game) ([]Game, error) {

	seen := map[string]struct{}{}
	depthJobs := jobs{make(map[int][]Game), sync.RWMutex{}}
	jobsCh := make(chan gameJob)
	errCh := make(chan error)

	game := g.Copy()
	seen[game.Hash()] = struct{}{}
	game.History = Instruction{}
	solutionsChan := make(chan Game)
	ctx, cancel := context.WithTimeout(context.Background(), b.MaxTime)
	defer cancel()
	solutions := []Game{}
	var err error
	var iterations = 1
	go func() {
		currentDepth := -1
		b.solveGame(ctx, game, jobsCh, solutionsChan, errCh, currentDepth, &g)
		currentDepth++
		for {
			if !depthJobs.has(currentDepth) {
				cancel()
				return
			}
			for {
				l := depthJobs.get(currentDepth)
				if l == nil {
					break
				}

				iterations++
				b.solveGame(ctx, *l, jobsCh, solutionsChan, errCh, currentDepth, &g)
			}
			currentDepth++
		}

	}()
	for {
		select {
		case error := <-errCh:
			if error == nil {
				continue
			}
			if s, ok := error.(SolverErr); ok {
				if s.ShouldQuit {
					err = error
					cancel()
				}
			}

		case job := <-jobsCh:
			if _, ok := seen[job.hash]; ok {
				continue
			}

			seen[job.hash] = struct{}{}
			if job.depth > b.MaxDepth {
				err = NewSolverErr(fmt.Errorf("Game-seen overflow (seen %d) (depth %d)", len(seen), job.depth), false)
				cancel()
				continue
			}
			// TODO: is there really a difference between this and the depth?
			if b.MaxMoves > 0 && b.MaxMoves < (g.Moves()-g.moves) {
				err = NewSolverErr(fmt.Errorf("Max-moves threshold triggered: %d, maxmoves %d", g.Moves(), b.MaxMoves), true)
				cancel()
				continue
			}
			if len(seen) > b.MaxVisits {
				err = NewSolverErr(fmt.Errorf("Game-seen overflow (seen %d) (depth %d)", len(seen), job.depth), false)
				cancel()
				continue
			}
			depthJobs.Lock()
			if depthJobs.queue[job.depth] == nil {
				depthJobs.queue[job.depth] = []Game{job.Game}
			} else {
				depthJobs.queue[job.depth] = append(depthJobs.queue[job.depth], job.Game)
			}
			depthJobs.Unlock()
		case solvedGame := <-solutionsChan:
			if b.MaxSolutions > 0 && len(solutions) >= b.MaxSolutions {
				cancel()
				continue
			}
			solutions = append(solutions, solvedGame)
			if solvedGame.Rules.GameMode == GameModeRandom && len(solutions) > 0 {
				if solvedGame.score-g.score > int64(b.InfiniteGameMaxScoreIncrease) {
					if err != nil && errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
						err = nil
					}
					return solutions, err
				}
			}
			if b.MaxSolutions > 0 && len(solutions) >= b.MaxSolutions {
				cancel()
			}
		case <-ctx.Done():
			contextErr := ctx.Err()
			if err != nil {
				return solutions, err
			}
			if contextErr != nil && len(solutions) > 0 && errors.Is(contextErr, context.DeadlineExceeded) || errors.Is(contextErr, context.Canceled) {
				contextErr = nil
			}
			return solutions, contextErr
		}
	}

}

// TODO: Major performance-boost is very much within reach with refactoring into a breadth-first implementation
// See https://github.com/runar-rkmedia/gotally/issues/14
func (b *bruteBreadthSolver) solveGame(
	ctx context.Context,
	g Game,
	jobsCh chan gameJob,
	solutions chan Game,
	errCh chan error,
	depth int,
	originalGame *Game,
) {
	depth++
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return
		}
		errCh <- NewSolverErr(fmt.Errorf("context: err %w", err), true)
		return
	}

	hints := g.GetHint()
	for _, h := range hints {
		gameCopy := g.Copy()
		ok := gameCopy.EvaluateForPath(h.Path)
		if !ok {
			errCh <- NewSolverErr(fmt.Errorf("Failed in game-solving for hint"), true)
			return
		}
		if gameCopy.IsGameWon() {
			solutions <- gameCopy
			continue
		}
		if gameCopy.Rules.GameMode == GameModeRandom {
			solutions <- gameCopy
		}
		hash := gameCopy.board.Hash()
		jobsCh <- gameJob{
			gameCopy, hash, depth, "hint",
		}
	}
	for _, dir := range []SwipeDirection{SwipeDirectionUp, SwipeDirectionRight, SwipeDirectionDown, SwipeDirectionLeft} {
		if !originalGame.Rules.NoReswipe && len(g.History) > 0 {
			// there is no point in swiping the same direction twice
			last := g.History[len(g.History)-1]
			if last == dir {
				continue
			}
			// there is no point in swiping the opposite direction of the last swipe
			// THIS IS ONLY TRUE FOR GAMES WHERE THERE IS NO Cell-generation
			// TODO: implement a check for this distinction
			if last == SwipeDirectionUp && dir == SwipeDirectionDown {
				continue
			}
			if last == SwipeDirectionDown && dir == SwipeDirectionUp {
				continue
			}
			if last == SwipeDirectionLeft && dir == SwipeDirectionRight {
				continue
			}
			if last == SwipeDirectionRight && dir == SwipeDirectionLeft {
				continue
			}
		}
		gameCopy := g.Copy()
		changed := gameCopy.Swipe(dir)
		if !changed {
			continue
		}
		hash := hex.EncodeToString([]byte(gameCopy.board.Hash()))
		jobsCh <- gameJob{
			gameCopy, hash, depth, "swipe",
		}
	}

}
