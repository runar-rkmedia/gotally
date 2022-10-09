<script lang="ts">
	export const ssr = false
	import 'pollen-css'
	import {
		GameMode,
		GetHintRequest,
		httpErrorStore,
		NewGameRequest,
		SwipeDirection
	} from '../connect-web'
	import { onMount } from 'svelte'
	import { browser } from '$app/env'
	import { animateSwipe } from '../logic'
	import type { PartialMessage } from '@bufbuild/protobuf/dist/types/message'
	import { ErrNoChange, store, storeHandler } from '../connect-web/store'
	import SwipeHint from '../components/board/SwipeHint.svelte'
	import GameWon from '../components/GameWon.svelte'
	import Dialog from '../components/Dialog.svelte'
	import CellComp from '../components/board/Cell.svelte'
	import { cellValue } from '../components/board/cell'
	import Counter from '../components/Counter.svelte'

	let boardDiv: HTMLDivElement

	const restartGame = () => {
		return storeHandler.commit(storeHandler.restartGame())
	}
	const getHint = async (options?: PartialMessage<GetHintRequest>) => {
		return storeHandler.commit(storeHandler.getHint(options))
	}
	const newGame = async (options: PartialMessage<NewGameRequest>) => {
		return storeHandler.commit(storeHandler.newGame(options))
	}

	onMount(async () => {
		await storeHandler.commit(storeHandler.getSession())

		// Set up swipes
		if (browser) {
			document.onkeydown = async (e) => {
				if ($store.didWin) {
					return
				}
				if (swipeLock) {
					e.preventDefault()
					return
				}
				if (selection.length) {
					e.preventDefault()
					return
				}
				switch (e.key) {
					case 'ArrowLeft':
					case 'a':
						swipe(SwipeDirection.LEFT)
						break
					case 'ArrowRight':
					case 'd':
						swipe(SwipeDirection.RIGHT)
						break
					case 'ArrowDown':
					case 's':
						swipe(SwipeDirection.DOWN)
						break
					case 'ArrowUp':
					case 'w':
						swipe(SwipeDirection.UP)
						break
					case 'h':
						getHint()
						break
					case 'c':
						restartGame()
						break
					case 'n':
						newGame({ mode: GameMode.RANDOM_CHALLENGE })
						break

					default:
						return
				}
				e.preventDefault()
			}
		}
	})
	const animateInvalidSwipe = (direction: SwipeDirection) => {
		console.error('Not implemented', 'animateInvalidSwipe', direction)
	}
	let swipeLock = false
	let _swipeQueueHandling = false
	const swipeQueue: SwipeDirection[] = []
	const swipe = async (direction: SwipeDirection) => {
		swipeQueue.push(direction)
		if (_swipeQueueHandling) {
			return
		}
		_swipeQueueHandling = true
		while (swipeQueue.length) {
			const dir = swipeQueue.pop()
			if (!dir) {
				_swipeQueueHandling = false
				return
			}
			await _swipe(dir)
		}
		_swipeQueueHandling = false
	}

	const _swipe = async (direction: SwipeDirection) => {
		if (!$store?.session?.game?.board) {
			return
		}
		if (selection.length) {
			return
		}
		if (swipeLock) {
			return
		}
		const swipeOptions = {
			swipeAnimationTime: 480,
			positive: direction === SwipeDirection.DOWN || direction === SwipeDirection.RIGHT,
			vertical: direction === SwipeDirection.UP || direction === SwipeDirection.DOWN,
			boardEl: boardDiv,
			nColumns: $store.session.game.board.columns,
			nRows: $store.session.game.board.rows
		}
		const shouldAnimate = await animateSwipe({ ...swipeOptions, dry: true })
		if (!shouldAnimate) {
			return
		}
		// swipeLock = true
		const r = storeHandler.swipe(direction)
		await animateSwipe(swipeOptions)
		// In case the server responds slower than the animation,
		// we set a visual indicator here.
		// On the other hand, if the server is faster, the should should see nothing of this
		boardDiv.style.opacity = '0.8'
		const [_, commit, err] = await r
		swipeLock = false
		for (const cell of [...boardDiv.children] as HTMLElement[]) {
			cell.style.transform = ''
			cell.style.transition = 'none'
		}
		boardDiv.style.opacity = '1'
		if (err) {
			if (err === ErrNoChange) {
				animateInvalidSwipe(direction)
			}
			return
		}
		commit()
	}
	function createSwiper(node: HTMLElement) {
		if (!browser) {
			return
		}
		let Hammer: any
		import('hammerjs').then((h) => {
			Hammer = h.default
			const hammerTime = new Hammer(node, {
				recognizers: [[Hammer.Swipe, { direction: Hammer.DIRECTION_ALL }]]
			})
			hammerTime.on('swipe', (e: any) => {
				switch (e.direction) {
					case Hammer.DIRECTION_UP:
						swipe(SwipeDirection.UP)
						break
					case Hammer.DIRECTION_DOWN:
						swipe(SwipeDirection.DOWN)
						break
					case Hammer.DIRECTION_LEFT:
						swipe(SwipeDirection.LEFT)
						break
					case Hammer.DIRECTION_RIGHT:
						swipe(SwipeDirection.RIGHT)
						break
				}
			})
		})
	}
	let selection: number[] = []
	let lastSelectionValue = 0
	let selectionMap: Record<number, boolean | undefined> = {}
	let invalidSelectionMap: Record<number, boolean | undefined> = {}
	const showSelectionInfo = true
	const select = async (i: number) => {
		invalidSelectionMap = {}
		const cell = $store.session.game.board.cells[i]
		if (!cell?.base) {
			invalidSelectionMap = {}
			selection = []
			lastSelectionValue = 0
			selectionMap = {}
			return
		}
		const isSelected = !!selectionMap[i]
		if (isSelected) {
			const [_, commit, err] = await storeHandler.combineCells(selection)
			if (err) {
				invalidSelectionMap = { [i]: true }
				selection = []
				lastSelectionValue = 0
				selectionMap = {}
				return
			}
			commit()
			selection = []
			selectionMap = {}
			lastSelectionValue = 0
			return
		}
		selection = [...selection, i]
		lastSelectionValue = cellValue(cell)
		selectionMap[i] = true
	}
	$: nextHint = $store.hintDoneIndex >= 0 ? $store.hints[$store.hintDoneIndex + 1] : $store.hints[0]
	$: selectionSum = !showSelectionInfo
		? 0
		: selection.reduce((r, i) => r + cellValue($store.session.game.board.cells[i]), 0)
	$: selectionProduct = !showSelectionInfo
		? 0
		: selection.reduce((r, i) => r * cellValue($store.session.game.board.cells[i]), 1)
</script>

{#if $httpErrorStore.errors.length}
	<div>
		{#each $httpErrorStore.errors as err}
			{err.error}
			<!-- content here -->
		{/each}
	</div>
	<!-- content here -->
{/if}

<div class="gameView">
	{#if $store?.session?.game?.board}
		<Dialog open={$store.didWin} let:open>
			<GameWon {open} />
		</Dialog>
		<div class="headControls">
			<div>
				<div class="score">
					Score: {$store.session.game.score}
				</div>
				<small class="boardName" title={$store.session.game.board.id}
					>{$store.session.game.board.name}</small
				>
				<div class="moves">
					Moves: {$store.session.game.moves}
				</div>
			</div>
		</div>

		<div class="boardContainer">
			<SwipeHint
				instruction={nextHint?.instructionOneof.value}
				active={nextHint?.instructionOneof.case === 'swipe'}
			/>
			<div
				bind:this={boardDiv}
				use:createSwiper
				class="board"
				style={`grid-template-columns: repeat(${$store.session.game.board.columns}, 1fr); grid-template-rows: repeat(${$store.session.game.board.rows}, 1fr)`}
			>
				{#each $store.session.game.board.cells as c, i}
					<CellComp
						noEval={invalidSelectionMap[i]}
						selected={selectionMap[i]}
						hinted={nextHint?.instructionOneof.case === 'combine' &&
							nextHint.instructionOneof.value.index.includes(i)}
						selectedLast={!!selection.length && selection[selection.length - 1] === i}
						cell={c}
						on:click={() => select(i)}
					/>
				{/each}
			</div>
		</div>
		{#if showSelectionInfo}
			<div class="selectionCounter">
				<Counter
					show={!!selectionSum}
					value={selectionSum}
					label="Sum"
					variant={lastSelectionValue * 2 < selectionSum
						? 'error'
						: lastSelectionValue * 2 === selectionSum
						? 'success'
						: 'normal'}
				/>
				<Counter
					show={selectionProduct > 1}
					value={selectionProduct}
					label="Product"
					variant={lastSelectionValue < selectionProduct / lastSelectionValue
						? 'error'
						: lastSelectionValue === selectionProduct / lastSelectionValue
						? 'success'
						: 'normal'}
				/>
			</div>
		{/if}
		<div class="bottom-controls">
			<button on:click={() => getHint()}>Hint </button>

			<div>
				<button disabled on:click={() => restartGame()}>Restart </button>
				<button on:click={() => newGame({ mode: GameMode.RANDOM })}>New Random game</button>
				<button disabled on:click={() => newGame({ mode: GameMode.TUTORIAL })}>New Tutorial</button>
				<button disabled on:click={() => newGame({ mode: GameMode.RANDOM_CHALLENGE })}
					>New Challenge</button
				>
			</div>
		</div>
	{/if}
</div>

<style>
	.gameView {
		height: 100%;
		max-height: 100%;
		display: flex;
		flex-direction: column;
	}

	.boardContainer {
		position: relative;
		border: 2px solid var(--color-blue-700);
		border-radius: var(--radius-lg);
		margin-block-end: var(--size-4);
		height: 100%;
		max-height: 100%;
	}
	.board {
		position: relative;
		transition: opacity 300ms var(--easing-standard);
		/* margin-inline: -4px; */
		display: grid;
		height: 100%;
	}
	.boardName {
		opacity: 0.7;
		float: right;
	}
	.selectionCounter {
		display: flex;
		justify-content: center;
		gap: 10px;
	}
	.bottom-controls {
		display: flex;
		justify-content: center;
		flex-direction: column;
	}
	button:disabled {
		opacity: 0.4;
	}
	button {
		background-color: var(--color-blue);
		transition: opacity 70ms var(--easing-standard);
		min-width: 52px;
		min-height: 52px;
		color: var(--color-black);
	}
</style>
